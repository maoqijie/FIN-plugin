module example-plugin

go 1.22

require (
	github.com/hashicorp/go-plugin v1.4.10
	github.com/maoqijie/FIN-plugin v0.0.0
)

replace github.com/maoqijie/FIN-plugin => ../../
