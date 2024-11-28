package route

import "github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"

var privateMiddleware = apiLayer.Middleware{PublicRoute: false}
