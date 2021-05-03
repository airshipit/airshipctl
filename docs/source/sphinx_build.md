## Steps to Sphinx Document Generation

This document outlines the steps to generate the html pages from the current .md or .rst files in this folder using sphinx build. These html pages can be rendered locally to get a view of how they would look in the https://docs.airshipit.org/airshipctl/ website finally. These steps can be used by the developer if he is making changes to files in this directory. The changed files then can be rendered locally using sphinx-build to get a view prior to merge.

Pre-requisite:

* Python3
* git
* go

Steps:

* Install sphinx: `python3 -m pip install sphinx`
* Clone airshipctl: `git clone https://github.com/airshipit/airshipctl.git`
* Make necessary changes to the documents in the folder docs/source
* If adding any new command to airshipctl generate documents files for that command using the below make command. This step would internally run `make cli-docs` which would generate the documents in docs/source/cli folder along with the required golden files. The user can run the `make cli-docs` command alternatively to generate the documents alone.
    `make update-golden`
* Build sphinx html pages: `cd docs/source; sphinx-build -b html . _build`
* Run local server: `cd _build; python3 -m http.server`
* Open URL to access the page: `http://localhost:8000/` navigate to the required section to validate the document(s) changed
