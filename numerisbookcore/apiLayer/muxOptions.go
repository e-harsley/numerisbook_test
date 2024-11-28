package apiLayer

import (
	"encoding/json"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/gorilla/mux"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"reflect"
	"strconv"
)

type MuxOption func(group *mux.Router) *mux.Router

type MuxOptions []MuxOption

func (co *MuxOptions) AddOptions(options ...MuxOption) {
	*co = append(*co, options...)
}

func muxOptions[M mongodb.DocumentModel](repository mongodb.MongoDb[M], option EndpointOption, serializer SerializerHandler) MuxOptions {
	var crud MuxOptions

	if option.allowList {
		fmt.Println("option.bindOption", option.bindOption)
		crud = append(crud, getList(repository, serializer.response, option.bindOption...))
	}
	if option.allowFetch {
		crud = append(crud, getOne(repository, serializer.response))
	}
	if option.allowCreate {
		crud = append(crud, create(repository, serializer.createDto, serializer.response, option.bindOption...))
	}
	if option.allowDelete {
		crud = append(crud, destroy(repository))
	}
	if option.allowUpdate {
		crud = append(crud, update(repository, serializer.createDto, serializer.response))
	}

	return crud
}

func getList[M mongodb.DocumentModel](repository mongodb.MongoDb[M], response interface{}, bindFrom ...string) MuxOption {
	fmt.Println("bindFrom getList", bindFrom)
	return func(router *mux.Router) *mux.Router {
		wrapper := func(w http.ResponseWriter, r *http.Request) {
			c := C{W: w, R: r}
			fmt.Println("helolo bindFrom", bindFrom)
			contextKeysToBind := extractContextBinds(bindFrom)
			paramMap := map[string]interface{}{}
			varsToBind := map[string]string{}
			fmt.Println("contextKeysToBind", contextKeysToBind)

			canBindAll := len(withoutContexts(bindFrom)) == 0

			if canBindAll || SliceContains(bindFrom, BindURI) {
				varsToBind = mux.Vars(r)
				fmt.Println("BindQuery BindQuery", BindQuery)
			}

			if canBindAll || SliceContains(bindFrom, BindQuery) {
				queryParams := r.URL.Query()
				for key, value := range queryParams {
					varsToBind[key] = value[0]
					//if primitiveUserID, err := primitive.ObjectIDFromHex(value[0]); err == nil {
					//	paramMap[key] = primitiveUserID
					//}
				}
			}

			fmt.Println("contextKeysToBind", paramMap)
			for _, cKey := range contextKeysToBind {
				fmt.Println("context", cKey)
				val := c.R.Context().Value(cKey)
				if val == nil {
					fmt.Println("warning!! context key == " + cKey + " does not exist")
					continue
				}

				jsonString, _ := json.Marshal(val)

				err := json.Unmarshal([]byte(jsonString), &paramMap)
				if err != nil {
					http.Error(w, fmt.Sprintf("context key %s made a bad cast: %v", cKey, err), http.StatusInternalServerError)
					return
				}
				for key, value := range paramMap {
					if str, ok := value.(string); ok {
						if primitiveUserID, err := primitive.ObjectIDFromHex(str); err == nil {
							fmt.Printf("Converted %s to ObjectID: %v\n", key, primitiveUserID)
							paramMap[key] = primitiveUserID
						}
					}
				}
				fmt.Println("dtoCastFrom", paramMap)
			}

			//err := bindContextAndQueryParams(r, &c, bindFrom, nil, &paramMap)
			//
			//if err != nil {
			//	http.Error(w, fmt.Sprintf("Failed to bind context and query params"), http.StatusInternalServerError)
			//	return
			//}

			param := c.Query("filter_by")

			pageSize, pageNumber := int64(10), int64(1)
			sortParam := c.Query("sort_by")
			pageSizeStr, pageNumberStr := c.Query("page_size"), c.Query("page")
			if pageSizeStr != "" {
				lim, _ := strconv.Atoi(pageSizeStr)
				pageSize = int64(lim)
			}
			if pageNumberStr != "" {
				page, _ := strconv.Atoi(pageNumberStr)
				pageNumber = int64(page)
			}
			skip := (pageNumber * pageSize) - pageSize

			var sortMap map[string]interface{}
			if sortParam == "" {
				sortMap = map[string]interface{}{"created_at": "desc"}
			} else {
				err := json.Unmarshal([]byte(sortParam), &sortMap)
				if err != nil {
					c.Response(400, err.Error(), "Failed to fetch")
					return
				}
			}
			fmt.Println("????? param", param)

			if param != "" {
				err := json.Unmarshal([]byte(param), &paramMap)
				fmt.Println(err)
				if err != nil {
					c.Response(400, err.Error(), "Failed to fetch")
					return
				}

			}
			fmt.Println("????? 2 param", param)
			params := buildQuery(paramMap)
			sorts := buildSort(sortMap)
			findOptions := mongodb.FindOptions{Limit: &pageSize, Skip: &skip, Sort: sorts}
			findOptions.SetFetchAllLinks(true)
			cursor, err := repository.Find(params, &findOptions)
			if err != nil {
				c.Response(400, err.Error(), "Failed to fetch")
				return
			}

			fmt.Println("cursor", cursor)

			var models []M

			err = cursor.ToSlice(&models)
			if err != nil {
				c.Response(400, err.Error(), "Failed to fetch")
				return
			}

			count, err := repository.Count(params)
			if err != nil {
				c.Response(400, err.Error(), "Failed to fetch")
				return
			}
			pagination := PaginationResponse{
				Page:    pageNumber,
				PerPage: pageSize,
				Skip:    skip,
				Count:   count,
			}

			if response == nil {
				c.GetResponse(http.StatusOK, params, sorts, pagination, &models, "FETCHED SUCCESSFULLY")
				return
			}

			resp, err := SerializerFunc(models, &response)
			if err != nil {
				c.Response(400, err.Error(), "Failed to fetch")
				return
			}
			c.GetResponse(http.StatusOK, params, sorts, pagination, resp, "FETCHED SUCCESSFULLY")
		}

		router.HandleFunc("", wrapper).Methods("GET")
		return router
	}
}

func getOne[M mongodb.DocumentModel](repository mongodb.MongoDb[M], response interface{}) MuxOption {
	return func(router *mux.Router) *mux.Router {
		wrapper := func(w http.ResponseWriter, r *http.Request) {
			c := C{W: w, R: r}
			id := c.Params("id")
			idHex, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.Response(400, err.Error(), "Invalid object id")
				return
			}
			fmt.Println(idHex)
			findOptions := mongodb.FindOptions{}
			findOptions.SetFetchAllLinks(true)

			res, err := repository.FindOne(bson.M{"_id": idHex}, &findOptions)

			if err != nil {
				c.Response(400, err.Error(), "VALIDATION ERROR")
				return
			}

			if response == nil {
				c.Response(http.StatusOK, res, "FETCHED SUCCESSFULLY")
				return
			}
			resp, err := SerializerFunc(res, response)
			if err != nil {
				c.Response(400, err.Error(), "FAILED TO FETCH")
				return
			}
			c.Response(http.StatusOK, resp, "FETCHED SUCCESSFULLY")
		}
		router.HandleFunc("/{id}", wrapper).Methods("GET")
		return router
	}
}

func create[M mongodb.DocumentModel](repository mongodb.MongoDb[M], schema interface{}, response interface{}, bindFrom ...string) MuxOption {
	return func(router *mux.Router) *mux.Router {
		wrapper := func(w http.ResponseWriter, r *http.Request) {
			var (
				err               error
				contextKeysToBind = extractContextBinds(bindFrom)
				varsToBind        = mux.Vars(r)
				canBindAll        = len(withoutContexts(bindFrom)) == 0
				c                 = C{W: w, R: r}
			)

			// Initialize schema if nil
			if schema == nil {
				var model M
				schema = model
			}

			// Create new instance of schema type
			schemaType := reflect.TypeOf(schema)
			if schemaType.Kind() == reflect.Ptr {
				schemaType = schemaType.Elem()
			}
			sche := reflect.New(schemaType).Interface()

			// Bind JSON
			if err := c.BindJSON(sche); err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "VALIDATION ERROR")
				return
			}

			// Handle query parameters
			if canBindAll || SliceContains(bindFrom, BindQuery) {
				queryParams := r.URL.Query()
				for key, value := range queryParams {
					if len(value) > 0 {
						varsToBind[key] = value[0]
					}
				}
			}
			fmt.Println(">>>>>>>", contextKeysToBind)
			// Handle context values
			ctx := r.Context()
			for _, cKey := range contextKeysToBind {
				val := ctx.Value(cKey)
				if val == nil {
					log.Printf("Warning: context key %s does not exist", cKey)
					continue
				}

				// Handle ObjectID conversion
				//if str, ok := val.(string); ok {
				//	if primitiveUserID, err := primitive.ObjectIDFromHex(str); err == nil {
				//		val = primitiveUserID
				//	}
				//}

				fmt.Println("here all ", reflect.TypeOf(val), val)

				// Create a temporary struct to hold the context value
				//tempStruct := reflect.New(schemaType).Interface()
				//if err := copier.Copy(tempStruct, val); err != nil {
				//	log.Printf("Error copying context value %s: %v", cKey, err)
				//	continue
				//}
				//fmt.Println("tempStruct", tempStruct)
				// Merge the temporary struct with the main schema
				if err := copier.Copy(sche, val); err != nil {
					http.Error(w, fmt.Sprintf("Failed to merge context key %s: %v", cKey, err), http.StatusInternalServerError)
					return
				}
			}

			// Validate if implements IRequest
			if validator, ok := sche.(IRequest); ok {
				if errRes := validator.Validate(); errRes != nil {
					resErr := ErrorDetail{}
					for key, value := range errRes.Errors {
						resErr = FormatValidationError(key, ValidateError, value)
					}

					w.Header().Set("Content-Type", "application/json")
					if jsonResponse, err := json.Marshal(resErr); err == nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write(jsonResponse)
						return
					}
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}
			}

			// Save to repository
			model, err := repository.Save(sche)
			if err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "FAILED TO SAVE")
				return
			}

			// Serialize response
			resp, err := SerializerFunc(model, &response)
			if err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "FAILED TO SERIALIZE")
				return
			}

			c.Response(http.StatusCreated, resp, "CREATED SUCCESSFULLY")
		}

		router.HandleFunc("", wrapper).Methods("POST")
		return router
	}
}

func update[M mongodb.DocumentModel](repository mongodb.MongoDb[M], schema interface{}, response interface{}, bindFrom ...string) MuxOption {
	return func(router *mux.Router) *mux.Router {
		wrapper := func(w http.ResponseWriter, r *http.Request) {

			var (
				err error

				varsToBind = mux.Vars(r)
				canBindAll = len(withoutContexts(bindFrom)) == 0

				c = C{W: w, R: r}
			)
			id := c.Params("id")
			idHex, err := primitive.ObjectIDFromHex(id)
			if err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "Invalid object id")
				return
			}
			if schema == nil {
				var mode M
				schema = mode
			}
			schemaType := reflect.TypeOf(schema)
			sche := reflect.New(schemaType).Interface()
			if err := c.BindJSON(sche); err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "VALIDATION ERROR")
				return
			}
			if canBindAll || SliceContains(bindFrom, BindQuery) {
				queryParams := r.URL.Query()
				for key, value := range queryParams {
					varsToBind[key] = value[0]
				}
			}
			for _, cKey := range varsToBind {
				val, exists := varsToBind[cKey]
				if !exists {
					fmt.Println("warning!! context key == " + cKey + " does not exist")
					continue
				}
				err := copier.Copy(&sche, val)
				if err != nil {
					http.Error(w, fmt.Sprintf("context key %s made a bad cast: %v", cKey, err), http.StatusInternalServerError)
					return
				}
			}
			valiSchema, ok := sche.(IRequest)
			if ok {
				errRes := valiSchema.Validate()
				if errRes != nil {
					fmt.Println("errRes >>>", errRes.Errors)
					res := errRes.Errors

					// Iterate over the map
					resErr := ErrorDetail{}

					for key, value := range res {
						resErr = FormatValidationError(key, ValidateError, value)
					}

					FormatErrRes(resErr, http.StatusBadRequest)

					w.Header().Set("Content-Type", "application/json")

					jsonResponse, jsonErr := json.Marshal(resErr)
					if jsonErr != nil {
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusBadRequest)
					w.Write(jsonResponse)
					return
				}
			}
			mode, err := repository.Update(bson.M{"_id": idHex}, &sche)
			if err != nil {
				c.Response(400, err.Error(), "FAILED TO SAVE")
				return
			}
			resp, err := SerializerFunc(mode, &response)
			if err != nil {
				c.Response(400, err.Error(), "FAILED TO SAVE")
				return
			}
			c.Response(http.StatusCreated, resp, "CREATED SUCCESSFULLY")
		}
		router.HandleFunc("/{id}", wrapper).Methods("PUT")
		return router
	}
}

func destroy[M mongodb.DocumentModel](repository mongodb.MongoDb[M]) MuxOption {
	return func(router *mux.Router) *mux.Router {
		wrapper := func(w http.ResponseWriter, r *http.Request) {
			c := C{W: w, R: r}
			objID := c.Params("id")
			id, err := primitive.ObjectIDFromHex(objID)
			if err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "Invalid object id")
				return
			}
			_, err = repository.Delete(bson.M{"_id": id})
			if err != nil {
				c.Response(http.StatusBadRequest, err.Error(), "FAILED TO SAVE")
				return
			}
			c.Response(http.StatusOK, map[string]interface{}{
				"message": fmt.Sprintf("%s deleted successfully", objID),
			}, "DELETED SUCCESSFULLY")
			return
		}
		router.HandleFunc("/{id}", wrapper).Methods("DELETE")
		return router
	}

}

func bindVarsToDTO[T any](dto *T, vars map[string]string) error {
	// Using reflection to set values on the struct
	v := reflect.ValueOf(dto).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !value.CanSet() {
			continue
		}

		if tag, ok := field.Tag.Lookup("json"); ok {
			if val, exists := vars[tag]; exists {
				if value.Kind() == reflect.String {
					value.SetString(val)
				} else if value.Kind() == reflect.Int {
					intVal, err := strconv.Atoi(val)
					if err == nil {
						value.SetInt(int64(intVal))
					}
				}
				// Add more type conversions as needed
			}
		}
	}

	return nil
}

func Depend[T IRequest](handler func(req T, c C) *Response, bindFrom ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			dtoCastFrom       T
			contextKeysToBind = extractContextBinds(bindFrom)
			varsToBind        map[string]string

			canBindAll = len(withoutContexts(bindFrom)) == 0

			c = C{W: w, R: r}
		)

		if canBindAll || SliceContains(bindFrom, BindJSON) {
			_ = json.NewDecoder(r.Body).Decode(&dtoCastFrom)

		}
		if canBindAll || SliceContains(bindFrom, BindURI) {
			varsToBind = mux.Vars(r)
		}

		if canBindAll || SliceContains(bindFrom, BindQuery) {
			queryParams := r.URL.Query()
			for key, value := range queryParams {
				varsToBind[key] = value[0]
			}
		}

		if err := bindVarsToDTO(&dtoCastFrom, varsToBind); err != nil {
			http.Error(w, fmt.Sprintf("Error binding variables: %v", err), http.StatusInternalServerError)
			return
		}

		for _, cKey := range contextKeysToBind {
			val := c.R.Context().Value(cKey)
			if val == nil {
				fmt.Println("warning!! context key == " + cKey + " does not exist")
				continue
			}
			fmt.Println(val)
			err := copier.Copy(&dtoCastFrom, val)
			if err != nil {
				http.Error(w, fmt.Sprintf("context key %s made a bad cast: %v", cKey, err), http.StatusInternalServerError)
				return
			}
			fmt.Println("dtoCastFrom", dtoCastFrom)
		}

		errRes := dtoCastFrom.Validate()
		if errRes != nil {
			fmt.Println("errRes >>>", errRes.Errors)
			res := errRes.Errors

			// Iterate over the map
			resErr := ErrorDetail{}

			for key, value := range res {
				resErr = FormatValidationError(key, ValidateError, value)
			}

			FormatErrRes(resErr, http.StatusBadRequest)

			w.Header().Set("Content-Type", "application/json")

			jsonResponse, jsonErr := json.Marshal(resErr)
			if jsonErr != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			w.Write(jsonResponse)
			return
		}

		res := handler(dtoCastFrom, c)
		if res == nil {
			http.Error(w, "No response in body", http.StatusInternalServerError)
			return
		}
		responseCode := res.HTTPStatusCode
		if responseCode == 0 {
			responseCode = MethodToStatusCode[r.Method]
			res.HTTPStatusCode = responseCode
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseCode)
		json.NewEncoder(w).Encode(res.JSONDATA())

	}
}
