package protocol

import (
	"errors"

	"github.com/capeprivacy/cli/entities"
	"github.com/gorilla/websocket"
)

type msg[T any] struct {
	Message T      `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeError(ws *websocket.Conn, err error) error {
	return ws.WriteJSON(msg[interface{}]{Error: err.Error()})
}

func writeMsg[T any](ws *websocket.Conn, req T) error {
	return ws.WriteJSON(msg[T]{Message: req})
}

func readMsg[T any](ws *websocket.Conn) (*T, error) {
	var msg msg[T]
	if err := ws.ReadJSON(&msg); err != nil {
		return nil, err
	}

	if msg.Error != "" {
		return nil, errors.New(msg.Error)
	}

	return &msg.Message, nil
}

type Hybrid interface {
	Decrypt(input []byte, contextInfo []byte) ([]byte, error)
}

type Protocol struct {
	Websocket *websocket.Conn
	Hybrid    Hybrid
}

func (p Protocol) Error(err error) error {
	return writeError(p.Websocket, err)
}

func (p Protocol) UserInput() ([]byte, error) {
	_, encInput, err := p.Websocket.ReadMessage()
	if err != nil {
		return nil, err
	}

	return p.Hybrid.Decrypt(encInput, nil)
}

func (p Protocol) WriteBinary(b []byte) error {
	return p.Websocket.WriteMessage(websocket.BinaryMessage, b)
}

func (p Protocol) WriteStart(request entities.StartRequest) error {
	return writeMsg(p.Websocket, request)
}

func (p Protocol) ReadStart() (*entities.StartRequest, error) {
	return readMsg[entities.StartRequest](p.Websocket)
}

// TODO -- `SetDeploymentIDRequest` is not a good type name
// they should be renamed to `DeploymentResponse` or something similar
func (p Protocol) WriteDeploymentResults(res entities.SetDeploymentIDRequest) error {
	return writeMsg(p.Websocket, res)
}

func (p Protocol) ReadDeploymentResults() (*entities.SetDeploymentIDRequest, error) {
	return readMsg[entities.SetDeploymentIDRequest](p.Websocket)
}

func (p Protocol) ReadRunResults() (*entities.RunResults, error) {
	return readMsg[entities.RunResults](p.Websocket)
}

func (p Protocol) WriteRunResults(results entities.RunResults) error {
	return writeMsg(p.Websocket, results)
}

func (p Protocol) ReadFunctionPublicKey() (*entities.FunctionPublicKey, error) {
	return readMsg[entities.FunctionPublicKey](p.Websocket)
}

func (p Protocol) WriteFunctionPublicKey(key string) error {
	return writeMsg(p.Websocket, entities.FunctionPublicKey{FunctionTokenPublicKey: key})
}

func (p Protocol) ReadAttestationDoc() ([]byte, error) {
	s, err := readMsg[entities.AttestationWrapper](p.Websocket)
	if err != nil {
		return nil, err
	}

	return s.Message, nil
}

func (p Protocol) WriteAttestationDoc(doc []byte) error {
	return writeMsg(p.Websocket, entities.AttestationWrapper{Type: "attestation_doc", Message: doc})
}
