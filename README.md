# Collage

> Collage (from the French: coller, "to glue"; French pronunciation: [kɔ.laʒ])
> is a technique of an art production, primarily used in the visual arts,
> where the artwork is made from an assemblage of different forms, thus
> creating a new whole.

Collage acts as a read-only [docker registry](https://github.com/docker/distribution)
that is made of images coming from different locations.

Users can define remote repositories hosted on different registries and "mount"
them under specific paths.


# Installation

Collage is a Go program, it can be installed by doing:

```
go get github.com/flavio/collage
```

You can also use the [flavio/collage](https://hub.docker.com/r/flavio/collage/)
image available on the Docker Hub.

```
$ docker run --rm -p 5000:5000 -v /path/to/local/config.json:/app/config.json:ro flavio/collage
```

**Note well:** this image is based on Alpine Linux. If you some of your registries
are using self-signed certificates you just have to mount them into the running
image under `/usr/local/share/ca-certificates/<cert.pem>`. The entrypoint of
the image will automatically load them.

openSUSE and SUSE packages are available inside of `obs://Virtualization:containers`.

# Usage

Collage expects a json configuration file that specifies all the mappings.

For example, given the following configuration file:

```json
{
  "mappings" : {
    "cool/stuff":  "index.docker.io/flavio",
    "cool/distro": "index.docker.io/opensuse",
    "etcd":        "quay.io/coreos/etcd",
    "foobar":      "http://insecure-registry.local.lan/foo"
  }
}
```

The repository `coreos/etcd` from the quay.io registry is mounted as repository
`etcd`.

The images of the repository `flavio/` from the Docker Hub are now available
under the `/cool/stuff` repository.
The same applies to the images of the `opensuse/` repository, now mounted as
`cool/distro`.

Having a collage instance running on `collage.local.lan`, the command
`docker pull collage.local.lan/etcd/etcd:v3.3` will retrieve
the `quay.io/coreos/etcd/etcd:v3.3` image transparently.

All the images inside of [/flavio](https://hub.docker.com/u/flavio/) are
available as `collage.local.lan/cool/stuff/<repository>:<tag(s)>`. For
example: [collage.local.lan/cool/stuff/guestbook](https://hub.docker.com/r/flavio/guestbook/)
[collage.local.lan/cool/stuff/guestbook](https://hub.docker.com/r/flavio/guestbook-go/),...

**Note well:** by default collage assumes the mapped registries are using TLS.
You must be explicit about registries **not** using TLS: provide the `http://`
prefix inside of their URL.

## Docker image

A Docker image is automatically built on the DockerHub based on the contents
of the `master` branch of this repository.

The image can be found [here](https://hub.docker.com/r/flavio/collage/), or
simply via:

```
docker pull flavio/collage
```

The image runs collage as unprivileged `web` user.

## Virtual hosts config

It's possible to configure collage to have virtual hosts specific mappings.

This can be achieved by using the following configuration:

```json
{
  "vhosts" : {
    "docker-io-mirror.local.lan": {
      "mappings" : {
        "/" : "mirror.local.lan/docker.io"
      }
    },
    "quay-io-mirror.local.lan": {
      "mappings" : {
        "/" : "mirror.local.lan/quay.io"
      }
    }
  },
  "mappings" : {
    "cool/stuff" : "index.docker.io/flavio",
    "cool/distro" : "index.docker.io/opensuse",
    "etcd": "quay.io/coreos/etcd"
    "foobar": "http://insecure-registry.local.lan/foo"
  }
}
```

Let's assume the host running collage can be reached using the following FQDNs:

  * `docker-io-mirror.local.lan`
  * `quay-io-mirror.local.lan`
  * `collage.local.lan`

That will lead to the following behaviours:

  * `docker pull docker-io-mirror.local.lan/busybox:latest` will
    be resolved to `mirror.local.lan/docker.io/busybox:latest`
  * `docker pull quay-io-mirror.local.lan/cores/etcd:latest` will
    be resolved to `mirror.local.lan/quay.io/coreos/etcd:latest`
  * `docker pull collage.local.lan/cool/stuff/collage:latest` will
    be resolved to `index.docker.io/flavio/collage:latest`

This is particularly useful when to host multiple "/" mappings with the same
collage instance.

**Note well:** also in this case collage assumes the mapped registries are using TLS.
You must be explicit about registries **not** using TLS: provide the `http://`
prefix inside of their URL.


# A nice use case

My team is building different docker images in an automated fashion. The images
are pushed to an internal registry (*registry.local.lan*) under a special
repository (`kubic/v3/`).

After an internal process the images are promoted to a production registry
(*registry.production.lan*) inside of a different repository (`/caasp`).

While developing we don't want to rewrite all our kubernetes manifest files,
helm charts, Dockerfile(s),... to point at `registry.local.lan/kubic/v3/...`.
For example we would like to keep referencing `velum:2.0` image as
`registry.production.lan/caasp/velum:2.0` instead of
`registry.local.lan/kubic/v3/velum:2.0`.

With collage we can setup a mapping like the following one:

```json
"mappings": {
  "caasp": "registry.local.lan/kubic/v3"
}
```

By doing that we can do a `docker pull collage.local.lan/caasp/velum:2.0`
and get our image downloaded.

We are almost there: the repository, name and tags are fine, but the registry
name is not (it's `collage.local.lan` instead of `registry.production.lan`).

## Enter docker mirroring

We have been working upstream to allow Docker to handle mirroring of 3rd party
registries. The work is ongoing [here](https://github.com/moby/moby/pull/34319).

With this patch in place we just have to configure `collage.local.lan` to
be used as a mirror of `registry.production.lan` and then we will be able
to perform:

```
docker pull registry.production.lan/caasp/velum:2.0
```

By doing that the docker engine will reach out to `collage.local.lan` looking
for `caasp/velum:2.0`. The collage instance will translate all the request
as redirections to `registry.local.lan/kubic/v3/velum:2.0`. No caching is
done by collage, but the final image will have the name we expect.

# Caveats

This is an experimental project, done for fun over a weekend. It surprisingly
works, but there are probably areas to make more robust and bugs ;)

Some known issues:

  * GET catalog: some registries don't allow this request to be performed
    (eg: docker hub, quay.io); there's nothing collage can do in that case.
    However everything works as expected if you know what you are looking for.
