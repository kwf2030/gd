# gd
a simple tool resolving go's dependency.

## Usage
`go get github.com/kwf2030/gd`

1. chdir to project root directory
2. create a vendor.json(example)
3. execute `gd`

## Internal Workflow
1. read vendor.json configuration
2. git clone/pull
3. git checkout(if specified a version)
4. copy dependencies to vendor directory
5. revert checkout to master(if checked out previously)
6. remove .git directory in vendors
