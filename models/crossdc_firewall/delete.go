package crossdc_firewall

type DeleteReq struct {
	DataCenter     string `valid:"required" URIParam:"yes"`
	FirewallPolicy string `valid:"required" URIParam:"yes"`
}
