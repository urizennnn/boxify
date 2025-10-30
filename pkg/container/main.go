package container

import (
	"fmt"
	"os"
)

func InitContainer() (error,string){
	err,containerID := CreateOverlayFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: error creating directory for upperDir %v\n", err)
		return err,""
	}
	return nil,containerID
}
