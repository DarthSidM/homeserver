package storage

import (
	"homeserver/internals/models"
	"log"
	"path/filepath"
	"syscall"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Agent struct {
	DB          *gorm.DB
	StorageRoot string
}

func (a *Agent) Run() {
	for {
		err := a.sync()
		if err != nil {
			log.Println("sync error:", err)
		}
		time.Sleep(10 * time.Second)
	}
}

func (a *Agent) sync() error {
	mounts, err := GetStorageMounts(a.StorageRoot)
	if err != nil {
		return err
	}

	activeMounts := map[string]bool{}

	for _, mount := range mounts {
		activeMounts[mount] = true

		total, used, err := getDiskUsage(mount)
		if err != nil {
			continue
		}

		name := filepath.Base(mount)
		diskUUID := name // assume mount folder name is UUID

		var storage models.Storage
		result := a.DB.Where("mount_path = ?", mount).First(&storage)

		if result.Error == gorm.ErrRecordNotFound {
			newStorage := models.Storage{
				ID:         uuid.New(),
				Name:       name,
				MountPath:  mount,
				DiskUUID:   diskUUID,
				TotalSpace: total,
				UsedSpace:  used,
				Status:     "active",
			}
			a.DB.Create(&newStorage)
			log.Println("Registered new storage:", mount)
		} else {
			a.DB.Model(&storage).Updates(models.Storage{
				TotalSpace: total,
				UsedSpace:  used,
				Status:     "active",
			})
		}
	}

	// mark missing as offline
	var all []models.Storage
	a.DB.Find(&all)

	for _, s := range all {
		if !activeMounts[s.MountPath] {
			a.DB.Model(&s).Update("status", "offline")
		}
	}

	return nil
}

func getDiskUsage(path string) (int64, int64, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, 0, err
	}

	total := int64(stat.Blocks) * int64(stat.Bsize)
	free := int64(stat.Bavail) * int64(stat.Bsize)
	used := total - free

	return total, used, nil
}
