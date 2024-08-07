package requests

import "net/http"

type configuration struct {
	retries         int
	headerDecorator HeaderDecorator
}

type Configurator func(c *configuration)

type HeaderDecorator func(header http.Header)

//goland:noinspection GoUnusedExportedFunction
func SetRetries(amount int) Configurator {
	return func(c *configuration) {
		c.retries = amount
	}
}

//goland:noinspection GoUnusedExportedFunction
func SetHeaderDecorator(hd HeaderDecorator) Configurator {
	return func(c *configuration) {
		c.headerDecorator = hd
	}
}
