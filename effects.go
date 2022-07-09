package nanoleaf

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// NanoEffects represents nanoleafs effects
type NanoEffects struct {
	nano     *Nanoleaf
	endpoint string
}

// EffectData effects data
type EffectData struct {
	Command              string         `json:"command,omitempty"`
	Loop                 bool           `json:"loop,omitempty"`
	Name                 string         `json:"animName,omitempty"`
	Type                 string         `json:"animType,omitempty"`
	Version              string         `json:"version,omitempty"`
	Data                 string         `json:"animData,omitempty"`
	ColorType            string         `json:"colorType,omitempty"`
	Palette              []PaletteColor `json:"palette,omitempty"`
	PluginType           string         `json:"pluginType,omitempty"`
	PluginUuid           string         `json:"pluginUuid,omitempty"`
	PluginOptions        []PluginOption `json:"pluginOptions,omitempty"`
	OverlayPalette       []PaletteColor `json:"overlayPalette,omitempty"`
	OverlayColorType     string         `json:"overlayColorType,omitempty"`
	HasOverlay           bool           `json:"hasOverlay,omitempty"`
	LogicalPanelsEnabled bool           `json:"logicalPanelsEnabled,omitempty"`
}

type PaletteColor struct {
	Hue         int     `json:"hue"`
	Saturation  int     `json:"saturation"`
	Brightness  int     `json:"brightness"`
	Probability float64 `json:"probability"`
}

type PluginOption struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"` // int, float, bool, or string
}

// newNanoEffects returns a new NanoEffects instance
func newNanoEffects(nano *Nanoleaf) *NanoEffects {
	return &NanoEffects{
		nano:     nano,
		endpoint: fmt.Sprintf("%s/%s/effects", nano.url, nano.token),
	}
}

// List lists all effects registered
func (e *NanoEffects) List() ([]string, error) {
	url := fmt.Sprintf("%s/effectsList", e.endpoint)
	resp, err := e.nano.client.R().Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return nil, ErrUnauthorized
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, ErrUnexpectedResponse
	}

	var effects []string

	if err := json.Unmarshal(resp.Body(), &effects); err != nil {
		return nil, ErrParsingJSON
	}

	return effects, nil
}

// Set sets given effects as active
func (e *NanoEffects) Set(name string) error {
	body := jsonPayload{"select": name}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusNoContent {
		return ErrUnexpectedResponse
	}

	return nil
}

// Get returns the currently active effect
func (e *NanoEffects) Get() (string, error) {
	url := fmt.Sprintf("%s/select", e.endpoint)
	resp, err := e.nano.client.R().Get(url)

	if err != nil {
		return "", err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}

	if resp.StatusCode() != http.StatusOK {
		return "", ErrUnexpectedResponse
	}

	var effect string

	if err := json.Unmarshal(resp.Body(), &effect); err != nil {
		return "", ErrParsingJSON
	}

	return effect, nil
}

// GetEffectData returns data of the given effect
func (e *NanoEffects) GetEffectData(effect string) (EffectData, error) {
	var data EffectData
	body := jsonPayload{
		"write": jsonPayload{
			"command":  "request",
			"animName": effect,
		},
	}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return data, err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return data, ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return data, ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusOK {
		return data, ErrUnexpectedResponse
	}

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		fmt.Println(err)
		return data, ErrParsingJSON
	}

	return data, nil
}

// returns data of all effects on the device
func (e *NanoEffects) GetAllEffectData() ([]EffectData, error) {

	animations := struct {
		Animations []EffectData `json:"animations"`
	}{}
	body := jsonPayload{
		"write": jsonPayload{
			"command": "requestAll",
		},
	}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return animations.Animations, err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return animations.Animations, ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return animations.Animations, ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusOK {
		return animations.Animations, ErrUnexpectedResponse
	}

	if err := json.Unmarshal(resp.Body(), &animations); err != nil {
		fmt.Println(err)
		return animations.Animations, ErrParsingJSON
	}

	return animations.Animations, nil
}

// /effects (rename)
func (e *NanoEffects) RenameEffect(animName, newName string) error {

	body := jsonPayload{
		"write": jsonPayload{
			"command":  "rename",
			"animName": animName,
			"newName":  newName,
		},
	}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusNoContent {
		return ErrUnexpectedResponse
	}

	return nil
}

// /effects (add/update)
func (e *NanoEffects) AddEffect(data EffectData, animName string) error {

	data.Command = "add"
	data.Name = animName
	body := jsonPayload{
		"write": data,
	}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusNoContent {
		return ErrUnexpectedResponse
	}

	return nil
}

// deletes the given effect
func (e *NanoEffects) DeleteEffect(animName string) error {

	body := jsonPayload{
		"write": jsonPayload{
			"command":  "delete",
			"animName": animName,
		},
	}
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode() == http.StatusNotFound {
		return ErrEffectNotFound
	}

	if resp.StatusCode() != http.StatusNoContent {
		return ErrUnexpectedResponse
	}

	return nil
}

// WriteRaw writes the raw command (outcome will depend on your body because the nanoleaf api is not well designed)
func (e *NanoEffects) WriteRaw(body jsonPayload) error {
	resp, err := e.nano.client.R().SetHeader("Content-Type", "application/json").SetBody(body).Put(e.endpoint)

	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode() != http.StatusNoContent {
		return ErrUnexpectedResponse
	}

	return nil
}

// displays effect with the given animData, looping if specified
func (e *NanoEffects) Display(data string, loop bool) error {
	body := jsonPayload{
		"write": jsonPayload{
			"command":   "display",
			"animType":  "custom",
			"colorType": "RGB",
			"animData":  data,
			"loop":      loop,
		},
	}

	return e.WriteRaw(body)
}

// temporarily displays the given effect for the specified duration
func (e *NanoEffects) DisplayTemp(animName string, duration int) error {
	body := jsonPayload{
		"write": jsonPayload{
			"command":  "displayTemp",
			"duration": duration,
			"animName": animName,
		},
	}

	return e.WriteRaw(body)
}

// ToString returns the effect as a string
func (e *NanoEffects) ToString(effect StreamEffect) string {
	data := fmt.Sprintf("%d", len(effect.Panels))

	for _, panel := range effect.Panels {
		data = fmt.Sprintf("%s %d %d", data, panel.ID, len(panel.Frames))

		for _, frame := range panel.Frames {
			data = fmt.Sprintf("%s %d %d %d 0 %d", data, frame.Red, frame.Green, frame.Blue, frame.Transition)
		}
	}

	return data
}
