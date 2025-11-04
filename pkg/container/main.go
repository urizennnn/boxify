package container

import (
	"log"
)

func InitContainer(containerID string) (error,string){
	err,containerID := CreateOverlayFS(containerID)
	if err != nil {
		log.Printf("Error: failed to create overlay %v\n", err)
		return err,""
	}
	return nil,containerID
}
