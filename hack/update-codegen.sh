client_gen=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)/generate-groups.sh
bash "$client_gen" \
     deepcopy,client \
     github.com/ian-howell/airshipctl/pkg/client \
     github.com/ian-howell/airshipctl/pkg/apis workflow:v1alpha1
cp -r "$GOPATH"/src/github.com/ian-howell/airshipctl/pkg "$(dirname "${0}")"/..
