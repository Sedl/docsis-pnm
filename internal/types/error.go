package types

type ApiError struct {
	ErrorStr string	`json:"error"`
	HttpStatusCode int `json:"-"`
}

func (err *ApiError) Error() string {
	return err.ErrorStr
}

/*
var ErrorInvalidMac = &ApiError{"Invalid MAC address", 400}
var ErrorConnectDatabase = &ApiError{"Can't connect to database. See logs for more information", 500}
var ErrorDbQuery = &ApiError{"Error while database query. See logs for more information", 500}
var ErrorModemNotFound = &ApiError{"No modem with this MAC", 404}
var ErrorCmtsNotFound = &ApiError{"No CMTS with the given hostname was found", 404}
 */