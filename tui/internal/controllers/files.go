package controllers

import (
	"fmt"
	"os"
	"path"

	"esdi/constants"

	"github.com/rivo/tview"
)

func buildFileNode(rootPath string, root *tview.TreeNode, files []os.DirEntry) {
	for _, f := range files {
		fullPath := path.Join(rootPath, f.Name())

		node := tview.NewTreeNode(f.Name())
		node.SetReference(fullPath)
		root.AddChild(node)

		if f.IsDir() {
			files, err := os.ReadDir(fullPath)
			// If we fail to read the directory we add an error message as that dir contents
			if err != nil {
				errNode := tview.NewTreeNode(fmt.Sprintf("error reading contents: %+v", err))
				node.AddChild(errNode)
				continue
			}

			buildFileNode(fullPath, node, files)
		}
	}
}

func buildFileTree(basePath string) (*tview.TreeNode, error) {
	root := tview.NewTreeNode(".")

	files, err := os.ReadDir(constants.CDashDisplayLayoutsDir)
	if err != nil {
		errNode := tview.NewTreeNode(fmt.Sprintf("error reading contents: %+v", err))
		root.AddChild(errNode)
	} else {
		buildFileNode(basePath, root, files)
	}

	return root, nil
}
