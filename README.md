# RainSFTP

RainSFTP is an implementaion of the [Secure File Transfer Protocol](https://en.wikipedia.org/wiki/SSH_File_Transfer_Protocol) backed by [LDAP](https://en.wikipedia.org/wiki/Lightweight_Directory_Access_Protocol) for authentication and an [S3](https://en.wikipedia.org/wiki/Amazon_S3)-compatible [object store](https://en.wikipedia.org/wiki/Object_storage).
This makes RainSFTP a powerful tool for managing cloud-based storage solutions where compatibility with SFTP clients is necessary.

Superficially, RainSFTP is similar to other object store-backed SFTP servers, but is designed to be as simple as possible and to only implement what is necessary to get users going.
It targets audiences where many SFTP servers need to be set up and torn down quickly in an enterprise environment where central authentication, central logging, audit trails, and other features are necessary.

## Features

- Compatible with most S3-compatible object stores, using the [Minio Go](https://github.com/minio/minio-go) client library
- Supports LDAP-based authentication and simple role-based access control
- Follows [12factor](https://12factor.net) design for easy deployment in cloud environments
- Generates logs in [Elastic Common Schema](https://www.elastic.co/guide/en/ecs/current/index.html) for easy processing
- Logs requests and responses, along with client IP addresses, to support a high quality [audit trail](https://en.wikipedia.org/wiki/Audit_trail)

## License

RainSFTP is released under the MIT license.
See [LICENSE.md](LICENSE.md).

## Etymology

RainSFTP is a product of the cloud.
Rain is a product of clouds.
Not super creative but it works.

## Limitations

- Files can be opened for reading or writing but not both due to difficulty finding an SFTP client that does this to test against
- `SETSTAT` is not supported as it is not particularly relevant or supported by S3-compatible environments
- `RENAME` is not supported currently as there is no rename operation in S3-compatible object stores, and instead a copy and delete operation will need to be implemented
- `LINK`, `SYMLINK` and `READLINK` are not supported as S3-compatible object stores do not have a concept of links

## Known Bugs

- Removing a non-empty directory will result in success even though the directory contents are not removed
