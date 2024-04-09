package wsclient

import (
	"encoding/json"
	"fmt"
)

type WsMsgData struct {
	Event	string			`json:"event"`
	Channel	string			`json:"channel"`
	Data	string			`json:"data"`
	Detail	string			`json:"detail"`
}

type WsMsgSubscribe struct {
	Subscribe	string		`json:"subscribe"`
}

type WsMsgClientStat struct {
	Mac		string		`json:"mac"`
	SiteId		string		`json:"site_id"`
	AssocTime	json.Number	`json:"assoc_time"`
	Family		string		`json:"family"`
	Model		string		`json:"model"`
	Os		string		`json:"os"`
	Manufacture	string		`json:"manufacture"`
	Bssid		string		`json:"bssid"`
	Username	string		`json:"username"`
	Hostname	string		`json:"hostname"`
	Ip		string		`json:"ip"`
	Ip6		string		`json:"ip6"`
	ApMac		string		`json:"ap_mac"`
	ApId		string		`json:"ap_id"`
	LastSeen	json.Number	`json:"last_seen"`
	Uptime		json.Number	`json:"uptime"`
	Ssid		string		`json:"ssid"`
	WlanId		string		`json:"wlan_id"`
	PskId		string		`json:"psk_id"`
	DualBand	bool		`json:"dual_band"`
	IsGuest		bool		`json:"is_guest"`
	KeyMgmt		string		`json:"key_mgmt"`
	Group		string		`json:"group"`
	Band		string		`json:"band"`
	Channel		json.Number	`json:"channel"`
	VlanId		string		`json:"vlan_id"`
	Proto		string		`json:"proto"`
	Rssi		json.Number	`json:"rssi"`
	Snr		json.Number	`json:"snr"`
	IdleTime	json.Number	`json:"idle_time"`
	TxRate		json.Number	`json:"tx_rate"`
	RxRate		json.Number	`json:"rx_rate"`
	TxPkts		json.Number	`json:"tx_pkts"`
	RxPkts		json.Number	`json:"rx_pkts"`
	TxBytes		json.Number	`json:"tx_bytes"`
	RxBytes		json.Number	`json:"rx_bytes"`
	TxRetries	json.Number	`json:"tx_retries"`
	RxRetries	json.Number	`json:"rx_retries"`
	TxBps		json.Number	`json:"tx_bps"`
	RxBps		json.Number	`json:"rx_bps"`
	MapId		string		`json:"map_id"`
	MapX		json.Number	`json:"x"`
	MapY		json.Number	`json:"y"`
	MapXM		json.Number	`json:"x_m"`
	MapYM		json.Number	`json:"y_m"`
	NumLocatingAps	json.Number	`json:"num_locating_aps"`
	Ttl		json.Number	`json:"_ttl"`
}

func (m *WsMsgClientStat) GetJsonKey(key string) (string, error) {
	switch key {
	case "mac":
		return m.Mac, nil
	case "site_id":
		return m.SiteId, nil
	case "assoc_time":
		return string(m.AssocTime), nil
	case "family":
		return m.Family, nil
	case "model":
		return m.Model, nil
	case "os":
		return m.Os, nil
	case "manufacture":
		return m.Manufacture, nil
	case "bssid":
		return m.Bssid, nil
	case "username":
		return m.Username, nil
	case "hostname":
		return m.Hostname, nil
	case "ip":
		return m.Ip, nil
	case "ip6":
		return m.Ip6, nil
	case "ap_mac":
		return m.ApMac, nil
	case "ap_id":
		return m.ApId, nil
	case "last_seen":
		return string(m.LastSeen), nil
	case "uptime":
		return string(m.Uptime), nil
	case "ssid":
		return m.Ssid, nil
	case "wlan_id":
		return m.WlanId, nil
	case "psk_id":
		return m.PskId, nil
	case "dual_band":
		return fmt.Sprintf("%v", m.DualBand), nil
	case "is_guest":
		return fmt.Sprintf("%v", m.IsGuest), nil
	case "key_mgmt":
		return m.KeyMgmt, nil
	case "group":
		return m.Group, nil
	case "band":
		return m.Band, nil
	case "channel":
		return string(m.Channel), nil
	case "vlan_id":
		return m.VlanId, nil
	case "proto":
		return m.Proto, nil
	case "rssi":
		return string(m.Rssi), nil
	case "snr":
		return string(m.Snr), nil
	case "idle_time":
		return string(m.IdleTime), nil
	case "tx_rate":
		return string(m.TxRate), nil
	case "rx_rate":
		return string(m.RxRate), nil
	case "tx_pkts":
		return string(m.TxPkts), nil
	case "rx_pkts":
		return string(m.RxPkts), nil
	case "tx_bytes":
		return string(m.TxBytes), nil
	case "rx_bytes":
		return string(m.RxBytes), nil
	case "tx_retries":
		return string(m.TxRetries), nil
	case "rx_retries":
		return string(m.RxRetries), nil
	case "tx_bps":
		return string(m.TxBps), nil
	case "rx_bps":
		return string(m.RxBps), nil
	case "map_id":
		return m.MapId, nil
	case "x":
		return string(m.MapX), nil
	case "y":
		return string(m.MapY), nil
	case "x_m":
		return string(m.MapXM), nil
	case "y_m":
		return string(m.MapYM), nil
	case "num_locating_aps":
		return string(m.NumLocatingAps), nil
	case "_ttl":
		return string(m.Ttl), nil
	default:
		return "", ErrNotFound
	}

	return "", ErrNotFound
}
