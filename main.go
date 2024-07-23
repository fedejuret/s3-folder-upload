package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	bucket := flag.String("bucket", "", "Nombre del bucket")
	bucketDestination := flag.String("bucket-destination", "", "Ruta de destino en el bucket")
	region := flag.String("region", "us-east-1", "Región de AWS")
	folder := flag.String("folder", "", "Ruta de la carpeta local")

	flag.Parse()

	if *bucket == "" || *bucketDestination == "" || *folder == "" {
		log.Fatal("Por favor, especifique todos los flags: -bucket, -bucket-destination y -folder")
	}

	if exists, err := exists(*folder); err != nil {
		log.Fatalf("Error al verificar la existencia de la carpeta: %v", err)
	} else if !exists {
		log.Fatalf("La carpeta %s no existe", *folder)
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*region),
	})
	if err != nil {
		log.Fatalf("Error al crear la sesión de AWS: %v", err)
	}

	uploader := s3manager.NewUploader(sess)

	err = filepath.Walk(*folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Error al acceder al archivo %s: %v", path, err)
		}

		relPath, err := filepath.Rel(*folder, path)
		if err != nil {
			return fmt.Errorf("Error al obtener la ruta relativa de %s: %v", path, err)
		}

		s3Key := filepath.Join(*bucketDestination, relPath)
		if !info.IsDir() {
			if err := uploadFile(uploader, *bucket, s3Key, path); err != nil {
				return fmt.Errorf("Error al subir %s a S3: %v", path, err)
			}
			fmt.Printf("Archivo subido exitosamente: %s\n", s3Key)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Error al iterar sobre los archivos: %v", err)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func uploadFile(uploader *s3manager.Uploader, bucket, key, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Error al abrir el archivo %s: %v", path, err)
	}
	defer file.Close()

	uploadInput := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	_, err = uploader.Upload(uploadInput)
	if err != nil {
		return fmt.Errorf("Error al subir %s a S3: %v", path, err)
	}
	return nil
}
