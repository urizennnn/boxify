package container

import (
	"log"
)

func InitContainer() (error,string){
	err,containerID := CreateOverlayFS()
	if err != nil {
		log.Printf("Error: failed to create overlay %v\n", err)
		return err,""
	}
	return nil,containerID
}
