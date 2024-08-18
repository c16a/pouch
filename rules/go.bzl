load("@rules_go//go:def.bzl", "go_binary")

def multi_arch_go_binary(name, targets, **kwargs):
    for target in targets:
        goos = target["goos"]
        goarch = target["goarch"]
        pure = target["pure"] if target["pure"] != None else "off"
        static = target["static"] if target["static"] != None else "auto"

        bin_name = "{0}_{1}_{2}".format(name, goos, goarch)
        go_binary(
            name = bin_name,
            gc_linkopts = [
                "-s",
                "-w",
            ],
            goarch = goarch,
            goos = goos,
            pure = pure,
            static = static,
            **kwargs
        )
