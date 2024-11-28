package service

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg/model"
)

var InvoiceConfigRepository = mongodb.NewMongoDb(model.InvoiceConfiguration{})
