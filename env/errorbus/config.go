package errorbus

import (
	"errors"
	"github.com/ottemo/foundation/env"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	config := env.GetConfig()
	if config == nil {
		return env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "6dee39acc930420eb777b95f6cab8981", "can't obtain config")
	}

	err := config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPatError,
		Value:       nil,
		Type:        env.ConstConfigItemGroupType,
		Editor:      "",
		Options:     nil,
		Label:       "Error",
		Description: "error handling settings",
		Image:       "",
	}, nil)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	// Log level
	logLevelValidator := func(newValue interface{}) (interface{}, error) {
		if newLevel, ok := newValue.(int); ok {
			if newLevel > 10 || newLevel < 0 {
				return logLevel, errors.New("'Log level' config value should be between 0 and 10")
			}
			logLevel = newLevel
		} else {
			return logLevel, errors.New("wrong type for 'Log level' config value")
		}
		return logLevel, nil
	}
	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathErrorLogLevel,
		Value:       5,
		Type:        "int",
		Editor:      "integer",
		Options:     nil,
		Label:       "Log level",
		Description: "errors below specified level will be send to logger service",
		Image:       "",
	}, logLevelValidator)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	// Hide level
	hideLevelValidator := func(newValue interface{}) (interface{}, error) {
		if newLevel, ok := newValue.(int); ok {
			if newLevel > 10 || newLevel < 0 {
				return hideLevel, errors.New("'Hide level' config value should be between 0 and 10")
			}
			hideLevel = newLevel
		} else {
			return hideLevel, errors.New("wrong type for 'Hide level' config value")
		}
		return hideLevel, nil
	}
	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathErrorHideLevel,
		Value:       5,
		Type:        "int",
		Editor:      "integer",
		Options:     nil,
		Label:       "Hide level",
		Description: "errors above specified level will be replaced to stub system error message",
		Image:       "",
	}, hideLevelValidator)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	// Hide message
	hideMessageValidator := func(newValue interface{}) (interface{}, error) {
		if newMessage, ok := newValue.(string); ok {
			hideMessage = newMessage
		} else {
			return hideMessage, errors.New("wrong type for 'Hide message' config value")
		}
		return hideMessage, nil
	}
	err = config.RegisterItem(env.StructConfigItem{
		Path:        ConstConfigPathErrorHideMessage,
		Value:       "System error has occured",
		Type:        "string",
		Editor:      "line_text",
		Options:     nil,
		Label:       "Hide message",
		Description: "system error message to substitute error message above hide level",
		Image:       "",
	}, hideMessageValidator)

	if err != nil {
		return env.ErrorDispatch(err)
	}

	return nil
}
