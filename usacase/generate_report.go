package usacase

import (
	"encoding/json"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/Vractos/go-reports-service/dto"
	"github.com/Vractos/go-reports-service/infra/kafka"
	"github.com/Vractos/go-reports-service/infra/repository"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func GenetareReport(requestJson []byte, repository repository.TransactionElasticRepository) error {
	var requestReport dto.RequestReport
	err := json.Unmarshal(requestJson, &requestReport)
	if err != nil {
		return err
	}
	data, err := repository.Search(requestReport.ReportID, requestReport.AccountID, requestReport.InitDate, requestReport.EndDate)
	if err != nil {
		return err
	}
	result, err := generateReportFile(data)
	if err != nil {
		return err
	}

	err = publishMessage(data.ReportID, string(result), "complete")
	if err != nil {
		return err
	}

	// err = os.Remove("data/" + data.ReportID + ".html")
	// if err != nil {
	// 	return err
	// }
	return nil
}

func generateReportFile(data dto.SearchResponse) ([]byte, error) {
	f, err := os.Create("data/" + data.ReportID + ".html")
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("report.html").ParseFiles("templates/report.html"))
	err = t.Execute(f, data)
	if err != nil {
		return nil, err
	}
	// result, err := uploadReport(data)
	// if err != nil {
	// 	return nil, err
	// }
	return []byte("www.google.com"), nil
}

func uploadReport(data dto.SearchResponse) (string, error) {
	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	fo, err := os.Open("data/" + data.ReportID + ".html")
	if err != nil {
		return "", err
	}
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("S3Bucket")),
		Key:    aws.String(data.ReportID + ".html"),
		Body:   fo,
	})

	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3Bucket")),
		Key:    aws.String(data.ReportID + ".html"),
	})

	if err != nil {
		return "", err
	}
	reportTTL, err := strconv.ParseInt(os.Getenv("ReportTTL"), 10, 64)
	if err != nil {
		return "", err
	}
	urlStr, err := req.Presign(time.Duration(reportTTL) * time.Hour)
	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func publishMessage(reportID string, fileURL string, status string) error {
	responseReport := dto.ResponseReport{
		ID:     reportID,
		FleURL: fileURL,
		Status: status,
	}

	responseJson, err := json.Marshal(responseReport)
	if err != nil {
		return err
	}
	producer := kafka.NewKafkaProducer()
	producer.SetupProducer(os.Getenv("KafkaBootstrapServers"))
	err = producer.Publish(string(responseJson), os.Getenv("KafkaProducerTopic"))
	if err != nil {
		return err
	}
	return nil
}
