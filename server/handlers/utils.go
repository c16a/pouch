package handlers

import (
	"crypto/tls"
	"github.com/c16a/pouch/server/store"
)

func GetTlsConfig(config *store.NodeConfig) (*tls.Config, error) {
	tlsAppConfig := config.Security.Tls
	if tlsAppConfig == nil || tlsAppConfig.Enable == false {
		return nil, nil
	} else {
		var cert tls.Certificate
		cert, err := tls.LoadX509KeyPair(tlsAppConfig.CertFilePath, tlsAppConfig.KeyFilePath)
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		return tlsConfig, err
	}
}
