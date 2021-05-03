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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"opendev.org/airship/airshipctl/cmd"
	"opendev.org/airship/airshipctl/pkg/fs"
)

var toctree = `####################
%s
####################

.. toctree::
   :maxdepth: 2

`

func main() {
	rootCmd := cmd.NewAirshipCTLCommand(os.Stdout)
	fs := fs.NewDocumentFs()
	dir := "./docs/source/cli"

	// Remote auto-generated notice
	rootCmd.DisableAutoGenTag = true

	// Generating .rst file for airshipctl root command
	if err := genReSTRoot(rootCmd, dir, fs); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}

	// Generating .rst files for airshipctl commands and sub-commands
	if err := genReST(rootCmd, dir, fs); err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}

// Generate .rst file for airshipctl root command
func genReSTRoot(cmd *cobra.Command, dir string, fs fs.FileSystem) error {
	basename := strings.Replace(cmd.CommandPath(), " ", "_", -1) + ".rst"
	fileName := filepath.Join(dir, basename)
	f, err := fs.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := fs.WriteFile(fileName, []byte(filePrepender(fileName))); err != nil {
		return err
	}

	// Invoking the doc.GenReSTCustom function to generate ReST file
	// for the airshipctl root command
	if err := doc.GenReSTCustom(cmd, f, linkHandler); err != nil {
		return err
	}
	return nil
}

// Default filePrepender
func filePrepender(filename string) string {
	return ""
}

// Custom linkHandler called when adding `SeeAlso` section
// to the document in the cobra.doc.GenReSTTreeCustom function
func linkHandler(name, ref string) string {
	return fmt.Sprintf(":ref:`%s <%s>`", name, ref)
}

// Generates .rst files for all the commands and subcommands in airshipctl
func genReST(cmd *cobra.Command, dir string, fs fs.FileSystem) error {
	// Used to populate the top level index.rst with each airshipctl command folder names
	fileName := filepath.Join(dir, "index.rst")
	if _, err := fs.Create(fileName); err != nil {
		return err
	}
	rootIndexRst := fmt.Sprintf(toctree, "Commands") + cmdFileName(cmd)

	// Loop over each airshipctl command like baremetal, phase, plan...etc
	for _, c := range cmd.Commands() {
		// Spliting the command to extract the command type
		// which would be baremetal, phase, plan...etc
		names := strings.Split(c.CommandPath(), " ")
		cmdName := names[len(names)-1]
		// Generates separate folder for current airshipctl command
		cmdDir := filepath.Join(dir, cmdName)
		if err := checkAndCreateDir(cmdDir, fs); err != nil {
			return nil
		}
		// Generating .rst files for all subcommands in the current airshipctl command
		if err := doc.GenReSTTreeCustom(c, cmdDir, filePrepender, linkHandler); err != nil {
			return err
		}
		// Create index.rst file for the current airshipctl command
		// This creates index.rst file in the appropriate folder
		// Example: In case of baremetal it creates an index.rst in the
		// baremetal folder. It adds the toctree details to file name.
		if err := genIndexReST(c, cmdDir, cmdName, fs); err != nil {
			return err
		}
		// Update the top-level index.rst file with the airshipctl command name
		// to invoke the index.rst in the sub folders
		// For example in case of baremetal: We create index.rst file for baremetal
		// subcommands in baremetal directory. So we refer this index.rst in the
		// top-level index.rst as below.
		rootIndexRst = rootIndexRst + "   " + cmdName + "/index\n"
	}
	if err := fs.WriteFile(fileName, []byte(rootIndexRst)); err != nil {
		return err
	}
	return nil
}

// Check the dir and if not present creates one
func checkAndCreateDir(dir string, fs fs.FileSystem) error {
	if !fs.Exists(dir) {
		if err := fs.Mkdir(dir); err != nil {
			return err
		}
	}
	if !fs.IsDir(dir) {
		return fmt.Errorf("expecting %s to be a directory", dir)
	}
	return nil
}

// Generates index.rst for a given airshipctl command
func genIndexReST(cmd *cobra.Command, cmdDir string, cmdName string, fs fs.FileSystem) error {
	cmdIndexFileName := filepath.Join(cmdDir, "index.rst")
	if _, err := fs.Create(cmdIndexFileName); err != nil {
		return err
	}
	cmdIndexRst := fmt.Sprintf(toctree, cmdName)
	// Updating the current index.rst with all the sub-commands file name details
	cmdIndexRst = updateIndexReSt(cmdIndexRst, cmd)
	if err := fs.WriteFile(cmdIndexFileName, []byte(cmdIndexRst)); err != nil {
		return err
	}
	return nil
}

// Updates the index.rst file with the subcommand file names
func updateIndexReSt(indexrst string, cmd *cobra.Command) string {
	indexrst += cmdFileName(cmd)
	for _, c := range cmd.Commands() {
		// Skipping help commands as they are not documented by the doc.cobra library
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		indexrst = updateIndexReSt(indexrst, c)
	}
	return indexrst
}

// Generate the file name corresponding to the cmd
func cmdFileName(cmd *cobra.Command) string {
	return "   " + strings.Replace(cmd.CommandPath(), " ", "_", -1) + "\n"
}
