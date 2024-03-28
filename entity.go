package main

type CreateRoomResponse struct {
	Code string `json:"code"`
}

type HTMXRequestHeaders struct {
	HXRequest     string `json:"HX-Request"`
	HXTrigger     string `json:"HX-Trigger"`
	HXTriggerName string `json:"HX-Trigger-Name"`
	HXTarget      string `json:"HX-Target"`
	HXCurrentURL  string `json:"HX-Current-URL"`
}
type HTMXRequest struct {
	Move    string             `json:"move"`
	Headers HTMXRequestHeaders `json:"HEADERS"`
}
