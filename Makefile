
install:
	go build -o ~/.steampipe/plugins/local/rackspace/rackspace.plugin *.go
	rm ~/.steampipe/logs/plugin*
	go build -o ~/.steampipe/plugins/local/rackspace/rackspace.plugin *.go

