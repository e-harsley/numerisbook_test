package apiLayer

import "fmt"

type EndpointOption struct {
	allowDelete bool
	allowList   bool
	allowFetch  bool
	allowCreate bool
	allowUpdate bool
	bindOption  []string
}

func defaultEndpointOpts() *EndpointOption {
	return &EndpointOption{
		allowList:   true,
		allowFetch:  true,
		allowCreate: true,
		allowDelete: false,
		allowUpdate: true,
	}
}

func WithoutList(opt *EndpointOption) {
	opt.allowList = false
}

func EndpointContextBind(bindContext string) EndpointOptFunc {
	fmt.Println("bindContext", bindContext)
	return func(opts *EndpointOption) {
		opts.bindOption = append(opts.bindOption, bindContext)
	}
}

func WithoutFetch(opt *EndpointOption) {
	opt.allowFetch = false
}

func WithoutCreate(opt *EndpointOption) {
	opt.allowCreate = false
}

func WithoutUpdate(opt *EndpointOption) {
	opt.allowUpdate = false
}

func WithDelete(opt *EndpointOption) {
	opt.allowDelete = true
}

type EndpointOptFunc func(*EndpointOption)
