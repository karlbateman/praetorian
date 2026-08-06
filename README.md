## Introduction

Praetorian is a zero dependency, lightweight service which wraps and unwraps
data encryption keys over HTTP using a rotating root key. It is intentionally
narrow in scope and uses the 256-bit Advanced Encryption Standard (AES) in
Galois Counter Mode (GCM). Its purpose is to provide a simplified method for
developers looking to implement [Envelope encryption], a technique used for
[encrypting sensitive data at rest].

Envelope encryption is where data is encrypted with a Data Encryption Key
(DEK), the DEK is then encrypted using a Key Encryption Key (KEK). The term
"envelope" refers to the process of wrapping one layer of encryption around
another, akin to sealing a letter with multiple layers of envelopes for added
security.

- **Data Encryption Key (DEK)**: A unique key is generated for each set of data
  to be encrypted. This key is used to encrypt and decrypt the data and should
  be symetric.
- **Key Encryption Key (KEK)**: A higher-level key used to encrypt and protect
  the DEK. The KEK is a symetric key stored securely in the Praetorian
  environment configuration.

Praetorian provides a rotating Key Encryption Key. This service provides two
HTTP endpoints so Data Encryption Keys can be wrapped and unwrapped, without
leaking any Key Encryption Key material.

[envelope encryption]: https://cloud.google.com/kms/docs/envelope-encryption
[encrypting sensitive data at rest]: https://cloud.google.com/docs/security/encryption/default-encryption

> [!NOTE]
> Throughout the remainder of this documentation, the term Key Encryption Key
> can be used interchangeably with Root Key.

## Why use Praetorian?

Praetorian is intended to be very easy to setup and great for startups or solo
developers looking to safely store PII and/or Special Category Data. In short,
if you're storing sensitive information in a database and want a simplified
approach to using envelope encryption for encryption-at-rest, you should use
Praetorian.

> [!NOTE]
> Praetorian does not use public key cryptography (asymetric encryption) and
> is therefor currently quantum safe. If you are using Praetorian as part of
> a system that uses or provides asymetric encryption, you should ensure you
> use PQC ([Post-Quantum Cryptography]). While the current symetric ciphers
> such as AES are believed to be relatively safe against quantum computers,
> there are ongoing efforts to develop more robust algorithms. As these become
> more broadly available and adopted, a future Praetorian release will include
> the standardised implementation to provide symetric PQC.

[post-quantum cryptography]: https://www.cloudflare.com/en-gb/learning/ssl/quantum/what-is-post-quantum-cryptography/

## Deployment

Before you start, you'll need to create a valid Praetorian configuration. Run
the following command from your local terminal to generate a JSON configuration
which contains a single root key that is set as the active root key.

```bash
echo $(printf '{"activeKeyId": "1", "rootKeys": {"1": "%s"}}' "$(openssl rand -base64 32)")
```

You can deploy a Praetorian service using the pre-built Docker Image available
at [karlbateman/praetorian]. When you deploy Praetorian you must have the
`PRAETORIAN_CONFIG` environment variable set with a JSON configuration
generated using the command above.

> [!CAUTION]
> Do not expose your Praetorian service to the public internet!

[karlbateman/praetorian]: https://hub.docker.com/r/karlbateman/praetorian

## Usage

Before integrating Praetorian into your application, it's best to start with a
locally running instance so you can explore and familiarise yourself with the
envelope encryption flow.

Assuming you have Docker installed, start by running the following command to
get a Praetorian container running and listening for requests on port `3000`.

```bash
docker run --rm \
  --name=praetorian \
  --publish="3000:3000" \
  --env=PRAETORIAN_CONFIG="$(printf '{"activeKeyId": "1", "rootKeys": {"1": "%s"}}' "$(openssl rand -base64 32)")" \
  karlbateman/praetorian
```

Now you have the Praetorian service running locally, you can issue a request to
wrap a data encryption key using the following `curl` request.

```bash
curl --silent \
  --request POST \
  --data '{"key": "abc123"}' \
  http://localhost:3000/wrap
```

To unwrap this data, simply copy the response from the terminal and send it to
the `/unwrap` endpoint.

```bash
curl --silent \
  --request POST \
  --data '<replace_with_wrap_response>' \
  http://localhost:3000/unwrap
```

This demonstrates the flow of wrapping and unwrapping data encryption keys.

## Operations

Praetorian is designed to be robust, performant and simple to operate. When
using Praetorian, there are two primary considerations.

### 1. Key Rotation

As a security best-practice, you should rotate your active root key at least
once every 3 months. This can be done by generating a new key on your local
machine and updating `"rootKeys"` in the `PRAETORIAN_CONFIG` to include the new
key. For example, if you have the following configuration:

```json
{
  "activeKeyId": "1",
  "rootKeys": {
    "1": "wS0J3KY/66FLxaGgZ3Djjl3sMRvqrIR1kjNKvgGZqtI="
  }
}
```

> [!CAUTION]
> Always keep a backup of your Praetorian configuration! A misconfiguration can
> result in not being able to decrypt previously encrypted data.

Then you would run the following command, to generate a new key:

```bash
openssl rand -base64 32
```

add the new key to `"rootKeys"` then update the `"activeKeyId"`.

```diff
 {
-  "activeKeyId": "1",
+  "activeKeyId": "2",
   "rootKeys": {
     "1": "wS0J3KY/66FLxaGgZ3Djjl3sMRvqrIR1kjNKvgGZqtI=",
+    "2": "UcCUyqWwP/Rv7/8c2zgO5gL1zQ/dTQjBMU84OwYYzN8="
   }
 }
```

Data encrypted with old keys can still be decrypted and when re-encrypted, will
be automatically updated to use the latest key. This process allows older keys
to be gracefully removed from circulation without causing any disruption.

### 2. Key Usage

To improve security and prevent misuse, the underlying AES256 encryption
implementation uses a managed nonce. Therefor, a root key MUST NOT be used to
encrypt more than 2^32 messages to limit the risk of random nonce collisions to
neglible levels.

This is fairly easy to mitigate if rotating your root key at a sufficient
cadence. For extremely high-traffic environments you can provision multiple
instances of Praetorian to divide the workload, using individual key rotation
policies based on traffic patterns.

## Roadmap

- [ ] Website
- [ ] Usage Examples
- [ ] Support Page
