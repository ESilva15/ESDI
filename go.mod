module esdi

go 1.22.0

require github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07

require golang.org/x/sys v0.26.0 // indirect

require esilva.org.localhost/bngsdk v0.1.0

replace esilva.org.localhost/bngsdk => ../pkg/bngsdk
