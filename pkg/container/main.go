package container

import (
	"fmt"
	"os"
)

func InitContainer() (error,string){
	err,containerID := CreateOverlayFS()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create overlay %v\n", err)
		return err,""
	}
	return nil,containerID
}
