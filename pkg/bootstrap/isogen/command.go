/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package isogen

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cheggaaa/pb/v3"

	"opendev.org/airship/airshipctl/pkg/api/v1alpha1"
	"opendev.org/airship/airshipctl/pkg/bootstrap/cloudinit"
	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/container"
	"opendev.org/airship/airshipctl/pkg/document"
	"opendev.org/airship/airshipctl/pkg/log"
	"opendev.org/airship/airshipctl/pkg/util"
)

const (
	builderConfigFileName = "builder-conf.yaml"

	// progressBarTemplate is a template string for progress bar
	// looks like 'Prefix [-->______] 20%' where Prefix is trimmed log line from docker container
	progressBarTemplate = `{{string . "prefix"}} {{bar . }} {{percent . }} `
	// defaultTerminalWidth is a default width of terminal if it's impossible to determine the actual one
	defaultTerminalWidth = 80
	// multiplier is a number of log lines produces while installing 1 package
	multiplier = 3
	// reInstallActions is a regular expression to check whether the log line contains of this substrings
	reInstallActions = `Extracting|Unpacking|Configuring|Preparing|Setting`
)

// Options is used for generate bootstrap ISO
type Options struct {
	CfgFactory config.Factory
	Progress   bool
}

// GenerateBootstrapIso will generate data for cloud init and start ISO builder container
// TODO (vkuzmin): Remove this public function and move another functions
// to the executor module when the phases will be ready
func (s *Options) GenerateBootstrapIso() error {
	ctx := context.Background()

	globalConf, err := s.CfgFactory()
	if err != nil {
		return err
	}

	root, err := globalConf.CurrentContextEntryPoint(config.BootstrapPhase)
	if err != nil {
		return err
	}
	docBundle, err := document.NewBundleByPath(root)
	if err != nil {
		return err
	}

	imageConfiguration := &v1alpha1.ImageConfiguration{}
	selector, err := document.NewSelector().ByObject(imageConfiguration, v1alpha1.Scheme)
	if err != nil {
		return err
	}
	doc, err := docBundle.SelectOne(selector)
	if err != nil {
		return err
	}

	err = doc.ToAPIObject(imageConfiguration, v1alpha1.Scheme)
	if err != nil {
		return err
	}
	if err = verifyInputs(imageConfiguration); err != nil {
		return err
	}

	log.Print("Creating ISO builder container")
	builder, err := container.NewContainer(
		&ctx, imageConfiguration.Container.ContainerRuntime,
		imageConfiguration.Container.Image)
	if err != nil {
		return err
	}

	err = createBootstrapIso(docBundle, builder, doc, imageConfiguration, log.DebugEnabled(), s.Progress)
	if err != nil {
		return err
	}
	log.Print("Checking artifacts")
	return verifyArtifacts(imageConfiguration)
}

func verifyInputs(cfg *v1alpha1.ImageConfiguration) error {
	if cfg.Container.Volume == "" {
		return config.ErrMissingConfig{
			What: "Must specify volume bind for ISO builder container",
		}
	}

	if (cfg.Builder.UserDataFileName == "") || (cfg.Builder.NetworkConfigFileName == "") {
		return config.ErrMissingConfig{
			What: "UserDataFileName or NetworkConfigFileName are not specified in ISO builder config",
		}
	}

	vols := strings.Split(cfg.Container.Volume, ":")
	switch {
	case len(vols) == 1:
		cfg.Container.Volume = fmt.Sprintf("%s:%s", vols[0], vols[0])
	case len(vols) > 2:
		return config.ErrInvalidConfig{
			What: "Bad container volume format. Use hostPath:contPath",
		}
	}
	return nil
}

func getContainerCfg(
	cfg *v1alpha1.ImageConfiguration,
	builderCfgYaml []byte,
	userData []byte,
	netConf []byte,
) map[string][]byte {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]

	fls := make(map[string][]byte)
	fls[filepath.Join(hostVol, cfg.Builder.UserDataFileName)] = userData
	fls[filepath.Join(hostVol, cfg.Builder.NetworkConfigFileName)] = netConf
	fls[filepath.Join(hostVol, builderConfigFileName)] = builderCfgYaml
	return fls
}

func verifyArtifacts(cfg *v1alpha1.ImageConfiguration) error {
	hostVol := strings.Split(cfg.Container.Volume, ":")[0]
	metadataPath := filepath.Join(hostVol, cfg.Builder.OutputMetadataFileName)
	_, err := os.Stat(metadataPath)
	return err
}

func createBootstrapIso(
	docBundle document.Bundle,
	builder container.Container,
	doc document.Document,
	cfg *v1alpha1.ImageConfiguration,
	debug bool,
	progress bool,
) error {
	cntVol := strings.Split(cfg.Container.Volume, ":")[1]
	log.Print("Creating cloud-init for ephemeral K8s")
	userData, netConf, err := cloudinit.GetCloudData(docBundle)
	if err != nil {
		return err
	}

	builderCfgYaml, err := doc.AsYAML()
	if err != nil {
		return err
	}

	fls := getContainerCfg(cfg, builderCfgYaml, userData, netConf)
	if err = util.WriteFiles(fls, 0600); err != nil {
		return err
	}

	vols := []string{cfg.Container.Volume}
	builderCfgLocation := filepath.Join(cntVol, builderConfigFileName)
	log.Printf("Running default container command. Mounted dir: %s", vols)

	envVars := []string{
		fmt.Sprintf("BUILDER_CONFIG=%s", builderCfgLocation),
		fmt.Sprintf("http_proxy=%s", os.Getenv("http_proxy")),
		fmt.Sprintf("https_proxy=%s", os.Getenv("https_proxy")),
		fmt.Sprintf("HTTP_PROXY=%s", os.Getenv("HTTP_PROXY")),
		fmt.Sprintf("HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY")),
		fmt.Sprintf("NO_PROXY=%s", os.Getenv("NO_PROXY")),
	}

	err = builder.RunCommand([]string{}, nil, vols, envVars)
	if err != nil {
		return err
	}

	log.Print("ISO generation is in progress. The whole process could take up to several minutes, please wait...")

	if debug || progress {
		var cLogs io.ReadCloser
		cLogs, err = builder.GetContainerLogs()
		if err != nil {
			log.Printf("failed to read container logs %s", err)
		} else {
			switch {
			case progress:
				showProgress(cLogs, log.Writer())
			case debug:
				log.Print("start reading container logs")
				// either container log output or progress bar will be shown
				if _, err = io.Copy(log.Writer(), cLogs); err != nil {
					log.Debugf("failed to write container logs to log output %s", err)
				}
				log.Print("got EOF from container logs")
			}
		}
	}

	if err = builder.WaitUntilFinished(); err != nil {
		return err
	}

	log.Print("ISO successfully built.")
	if !debug {
		log.Print("Removing container.")
		return builder.RmContainer()
	}

	log.Debugf("Debug flag is set. Container %s stopped but not deleted.", builder.GetID())
	return nil
}

func showProgress(reader io.ReadCloser, writer io.Writer) {
	reFindActions := regexp.MustCompile(reInstallActions)

	var bar *pb.ProgressBar

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	// Reading container log line by line
	for scanner.Scan() {
		curLine := scanner.Text()
		// Trying to find entry points of package installation
		switch {
		case strings.Contains(curLine, "Retrieving Packages") ||
			strings.Contains(curLine, "newly installed"):
			finalizePb(bar, "Completed")

			pkgCount := calculatePkgCount(scanner, writer, curLine)
			if pkgCount > 0 {
				bar = pb.ProgressBarTemplate(progressBarTemplate).Start(pkgCount * multiplier)
				bar.SetWriter(writer)
				setPbPrefix(bar, "Installing required packages")
			}
		case strings.Contains(curLine, "Base system installed successfully") ||
			strings.Contains(curLine, "mksquashfs"):
			finalizePb(bar, "Completed")

		case bar != nil && bar.IsStarted():
			if reFindActions.MatchString(curLine) {
				if bar.Current() < bar.Total() {
					setPbPrefix(bar, curLine)
					bar.Increment()
				}
			}
		case strings.Contains(curLine, "filesystem.squashfs"):
			fmt.Fprintln(writer, curLine)
		}
	}

	finalizePb(bar, "An unexpected error occurred while log parsing")
}

func finalizePb(bar *pb.ProgressBar, msg string) {
	if bar != nil && bar.IsStarted() {
		bar.SetCurrent(bar.Total())
		setPbPrefix(bar, msg)
		bar.Finish()
	}
}

func setPbPrefix(bar *pb.ProgressBar, msg string) {
	terminalWidth := defaultTerminalWidth
	halfWidth := terminalWidth / 2
	bar.SetWidth(terminalWidth)
	if len(msg) > halfWidth {
		msg = fmt.Sprintf("%v...", msg[0:halfWidth-3])
	} else {
		msg = fmt.Sprintf("%-*v", halfWidth, msg)
	}
	bar.Set("prefix", msg)
}

func calculatePkgCount(scanner *bufio.Scanner, writer io.Writer, curLine string) int {
	reFindNumbers := regexp.MustCompile("[0-9]+")

	// Trying to count how many packages is going to be installed
	pkgCount := 0
	matches := reFindNumbers.FindAllString(curLine, -1)
	if matches == nil {
		// There is no numbers is line about base packages, counting them manually to get estimates
		fmt.Fprint(writer, "Retrieving base packages ")
		for scanner.Scan() {
			curLine = scanner.Text()
			if strings.Contains(curLine, "Retrieving") {
				pkgCount++
				fmt.Fprint(writer, ".")
			}
			if strings.Contains(curLine, "Chosen extractor") {
				fmt.Fprintln(writer, " Done")
				return pkgCount
			}
		}
	}
	if len(matches) >= 2 {
		for _, v := range matches[0:2] {
			j, err := strconv.Atoi(v)
			if err != nil {
				continue
			}
			pkgCount += j
		}
	}
	return pkgCount
}
