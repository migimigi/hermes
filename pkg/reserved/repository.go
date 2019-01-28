package reserved

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/itsubaki/hermes/pkg/awsprice"
)

type Repository struct {
	Profile  string     `json:"profile"`
	Region   []string   `json:"region"`
	Internal RecordList `json:"internal"`
}

func NewRepository(profile string, region []string) (*Repository, error) {
	repo := &Repository{
		Profile: profile,
		Region:  region,
	}

	for _, r := range region {
		ses, err := session.NewSession(&aws.Config{Region: aws.String(r)})
		if err != nil {
			return nil, fmt.Errorf("new session (region=%s): %v", r, err)
		}

		{
			output, err := ec2.New(ses).DescribeReservedInstances(&ec2.DescribeReservedInstancesInput{
				Filters: []*ec2.Filter{
					{Name: aws.String("state"), Values: []*string{aws.String("active")}},
				},
			})
			if err != nil {
				return nil, fmt.Errorf("describe reserved instances: %v", err)
			}

			for _, i := range output.ReservedInstances {
				repo.Internal = append(repo.Internal, &Record{
					Region:             r,
					Duration:           *i.Duration,
					OfferingType:       *i.OfferingType,
					OfferingClass:      *i.OfferingClass,
					ProductDescription: *i.ProductDescription,
					InstanceType:       *i.InstanceType,
					InstanceCount:      *i.InstanceCount,
					Start:              *i.Start,
				})
			}
		}

		{
			client := elasticache.New(ses)
			var maker *string
			for {
				input := &elasticache.DescribeReservedCacheNodesInput{}
				if maker != nil {
					input.Marker = maker
				}

				output, err := client.DescribeReservedCacheNodes(input)
				if err != nil {
					return nil, fmt.Errorf("describe reserved cachenode: %v", err)
				}

				for _, i := range output.ReservedCacheNodes {
					if *i.State != "active" {
						continue
					}

					repo.Internal = append(repo.Internal, &Record{
						Region:             r,
						Duration:           *i.Duration,
						OfferingType:       *i.OfferingType,
						ProductDescription: *i.ProductDescription,
						CacheNodeType:      *i.CacheNodeType,
						CacheNodeCount:     *i.CacheNodeCount,
						Start:              *i.StartTime,
					})
				}

				if maker == nil {
					break
				}
			}
		}

		{
			client := rds.New(ses)
			var maker *string
			for {
				input := &rds.DescribeReservedDBInstancesInput{}
				if maker != nil {
					input.Marker = maker
				}

				output, err := client.DescribeReservedDBInstances(input)
				if err != nil {
					return nil, fmt.Errorf("describe reserved db instance: %v", err)
				}

				for _, i := range output.ReservedDBInstances {
					if *i.State != "active" {
						continue
					}

					repo.Internal = append(repo.Internal, &Record{
						Region:             r,
						Duration:           *i.Duration,
						OfferingType:       *i.OfferingType,
						ProductDescription: *i.ProductDescription,
						DBInstanceClass:    *i.DBInstanceClass,
						DBInstanceCount:    *i.DBInstanceCount,
						Start:              *i.StartTime,
						MultiAZ:            *i.MultiAZ,
					})
				}

				if maker == nil {
					break
				}
			}
		}
	}

	return repo, nil
}

func Read(path string) (*Repository, error) {
	read, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}

	repo := &Repository{}
	if err := repo.Deserialize(read); err != nil {
		return nil, fmt.Errorf("new repository: %v", err)
	}

	return repo, nil
}

func (r *Repository) Write(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return nil
	}

	bytes, err := r.Serialize()
	if err != nil {
		return fmt.Errorf("serialize: %v", err)
	}

	if err := ioutil.WriteFile(path, bytes, os.ModePerm); err != nil {
		return fmt.Errorf("write file: %v", err)
	}

	return nil
}

func (r *Repository) Serialize() ([]byte, error) {
	bytes, err := json.Marshal(r)
	if err != nil {
		return []byte{}, fmt.Errorf("marshal: %v", err)
	}

	return bytes, nil
}

func (r *Repository) Deserialize(bytes []byte) error {
	if err := json.Unmarshal(bytes, r); err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}

	return nil
}

func (r *Repository) SelectAll() RecordList {
	return r.Internal
}

func (r *Repository) FindByAWSPrice(record *awsprice.Record) (*Record, error) {
	duration := 31536000
	if record.LeaseContractLength == "3yr" {
		duration = 94608000
	}

	rs := r.SelectAll().
		Region(record.Region).
		InstanceType(record.InstanceType).
		Duration(int64(duration)).
		OfferingClass(record.OfferingClass).
		OfferingType(record.PurchaseOption).
		ContainsProductDescription(record.OperatingSystem)

	if len(rs) < 1 {
		return nil, fmt.Errorf("not found")
	}

	if len(rs) > 1 {
		return nil, fmt.Errorf("invalid query to awsprice repository")
	}

	return rs[0], nil
}
