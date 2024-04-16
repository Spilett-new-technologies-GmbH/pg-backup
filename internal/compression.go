package internal

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

// CompressBackup generates compressed file with the same name in the same folder with gz encryption
func CompressBackup(backupFilePath string) error {
	// Open the original backup file
	backupFile, err := os.Open(backupFilePath)
	if err != nil {
		return fmt.Errorf("error opening backup file: %v", err)
	}
	defer backupFile.Close()

	// Create the compressed backup file
	compressedBackupFile, err := os.Create(backupFilePath + ".gz")
	if err != nil {
		return fmt.Errorf("error creating compressed backup file: %v", err)
	}
	defer compressedBackupFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(compressedBackupFile)
	defer gzipWriter.Close()

	// Copy the contents of the original backup file to the gzip writer
	_, err = io.Copy(gzipWriter, backupFile)
	if err != nil {
		return fmt.Errorf("error compressing backup file: %v", err)
	}

	fmt.Println("Backup compressed successfully:", backupFilePath+".gz")
	return nil
}
