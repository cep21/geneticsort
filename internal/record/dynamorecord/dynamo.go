package dynamorecord

import (
	"context"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cep21/geneticsort/internal/record"
)

type Recorder struct {
	Client    *dynamodb.DynamoDB
	TableName string
}

var _ record.Recorder = &Recorder{}

func (d *Recorder) Record(ctx context.Context, r record.Record) error {
	_, err := d.Client.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: &d.TableName,
		Item: map[string]*dynamodb.AttributeValue{
			"key": {
				S: aws.String(r.Hash()),
			},
			"best": {
				S: aws.String(r.BestCandidate.String()),
			},
			"fitness": {
				N: aws.String(strconv.Itoa(r.BestCandidate.Fitness())),
			},
			"parent_select": {
				S: aws.String(r.Algorithm.ParentSelector.String()),
			},
			"family": {
				S: aws.String(r.Algorithm.Factory.Family()),
			},
			"mutator": {
				S: aws.String(r.Algorithm.Mutator.String()),
			},
			"terminator": {
				S: aws.String(r.Algorithm.Terminator.String()),
			},
			"crossover": {
				S: aws.String(r.Algorithm.Crossover.String()),
			},
			"survivor_selection": {
				S: aws.String(r.Algorithm.SurvivorSelection.String()),
			},
			"population_size": {
				N: aws.String(strconv.Itoa(r.Algorithm.PopulationSize)),
			},
		},
	})
	return err
}
