load("@rules_go//go:def.bzl", "go_binary")
load("@rules_oci//oci:defs.bzl", "oci_image")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")

def multi_arch_go_image(name, targets, **kwargs):
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
        pkg_tar(
            name = "tar_{0}".format(bin_name),
            srcs = [":{0}".format(bin_name)],
        )
        pkg_tar(
            name = "certs_tar_{0}".format(bin_name),
            srcs = ["@mozilla_cert_pem//file"],
            package_dir = "/etc/ssl/certs",
        )
        oci_image(
            name = "image_{0}".format(bin_name),
            os = goos,
            architecture = goarch,
            entrypoint = ["/{0}".format(bin_name)],
            tars = [":tar_{0}".format(bin_name), ":certs_tar_{0}".format(bin_name)],
        )
