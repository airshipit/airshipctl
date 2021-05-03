#!/usr/bin/python3

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from os import getenv
from os.path import isfile, join

from openapi2jsonschema import command
from openapi2jsonschema.command import debug, info, error
from openapi_spec_validator import validate_v3_spec
from openapi_spec_validator.exceptions import OpenAPIValidationError
from ruamel.yaml import YAML

openapi_schema_path = "/workdir/schemas-cache/openapischema"
openapi_schema_pattern = """openapi: 3.0.0
info:
  title: title
  version: 1.0.1
paths: {}
components:
  schemas: {}
"""
crd_kind = "CustomResourceDefinition"
crd_list = "/workdir/schemas-cache/crd-list"
rewrite_env = "VALIDATOR_REWRITE_SCHEMAS"
yaml = YAML()


def get_gvk(crd):
    """ Extracts group, version(s), kind data from CRD """
    group = crd["spec"]["group"].split(".")[0]
    kind = crd["spec"]["names"]["kind"].lower()

    try:
        version = crd["spec"]["version"]  # v1beta1 CRD
    except KeyError:
        version = crd["spec"]["versions"]  # v1 CRD

    return group, version, kind


def process_crd(crd, schemas, schemas_location, rewrite=False):
    """ Processes CRD document, extracts GVK and corresponding OpenAPIV3Schema(s) """
    g, v, k = get_gvk(crd)  # get GVK as tuple

    if isinstance(v, str):  # process CRD as v1beta1
        try:
            gvk = g + '.' + v + '.' + k
            kgv = k + "-" + g + "-" + v + ".json"

            # do not rewrite schemas by default if already exists
            if (not isfile(join(schemas_location, kgv)) and gvk not in schemas) or rewrite:
                schemas[gvk] = crd["spec"]["validation"]["openAPIV3Schema"]
                debug("Extracting OpenAPIV3Schema for {}".format(gvk))
            else:
                debug("OpenAPIV3Schema for {} was already processed, skipping".format(gvk))
        except KeyError:
            error("Cannot find OpenAPIV3Schema for {}".format(k))
        return

    if isinstance(v, list):  # process CRD as v1
        for version in v:
            try:
                gvk = g + '.' + version["name"] + '.' + k
                kgv = k + "-" + g + "-" + version["name"] + ".json"

                # do not rewrite schemas by default if already exists
                if (not isfile(join(schemas_location, kgv)) and gvk not in schemas) or rewrite:
                    schemas[gvk] = version["schema"]["openAPIV3Schema"]
                    debug("Extracting OpenAPIV3Schema for {}".format(gvk))
                else:
                    debug("OpenAPIV3Schema for {} was already processed, skipping".format(gvk))
            except KeyError:
                error("Cannot find OpenAPIV3Schema for {}".format(k))
                continue
        return


def check_yaml_kind(data):
    """ Determines whether a YAML document has CRD kind """
    return True if data is not None and "kind" in data and data["kind"] == crd_kind else False


def run():
    """
    The main function. Reads CRDs from URLs, intelligently extracts OpenAPIV3Schema(s)
    from each CRD, appends OpenAPIV3Schema to a designated file and verifies through OpenAPIValidator
    """
    openapi_schema = yaml.load(openapi_schema_pattern)
    schemas = openapi_schema["components"]["schemas"]

    with open(crd_list, 'r') as crd_list_file:  # read file with CRD locations
        crd_list_data = yaml.load(crd_list_file)

    with open(crd_list_data['crdList'], 'r') as yaml_file:
        crd_data = yaml.load_all(yaml_file)  # read CRDs
        for crd in crd_data:
            try:
                if check_yaml_kind(crd):
                    process_crd(crd, schemas, crd_list_data["schemasLocation"], getenv(rewrite_env) is not None)
            except Exception as exc:
                error("An error occurred while processing CRD data from phase rendered docs\n{}".format(exc))

    # Validate output V3 spec
    try:
        validate_v3_spec(openapi_schema)
        info("Validation of OpenAPIV3Schemas is successful")
    except OpenAPIValidationError as exc:
        error("An error occurred while validating OpenAPIV3Schema")
        raise exc

    # Rewrite openAPI schema file
    with open(openapi_schema_path, 'w') as openapi_schema_file:
        info("Saving OpenAPIV3Schemas")
        yaml.dump(openapi_schema, openapi_schema_file)

    # run openapi2jsonschema conversion
    command.default()


if __name__ == "__main__":
    run()
