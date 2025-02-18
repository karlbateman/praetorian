## Introduction

Praetorian is a zero dependency, lightweight service which wraps and unwraps
data encryption keys over HTTP using a rotating root key. It is intentionally
narrow in scope and uses the 256-bit Advanced Encryption Standard (AES) in
Galois Counter Mode (GCM). Its purpose is to provide a simplified method for
developers looking to implement [Envelope encryption], a technique used to
[encrypting sensitive data at rest].

Envelope encryption is a technique where data is encrypted with a Data
Encryption Key (DEK), the DEK is encrypted using a Key Encryption Key (KEK). The
term "envelope" refers to the process of wrapping one layer of encryption around
another, akin to sealing a letter with multiple layers of envelopes for added
security.

- **Data Encryption Key (DEK)**: A unique key is generated for each set of data
  to be encrypted. This key is used to encrypt and decrypt the data and should
  be symetric.
- **Key Encryption Key (KEK)**: A higher-level key used to encrypt and protect
  the DEK. The KEK is an asymetric key stored securely in the Praetorian
  environment configuration.

Praetorian provides the Key Encryption Key which can be rotated and exposes HTTP
endpoints so Data Encryption Keys can be wrapped and unwrapped, without leaking
any Key Encryption Key material.

[envelope encryption]: https://cloud.google.com/kms/docs/envelope-encryption
[encrypting sensitive data at rest]: https://cloud.google.com/docs/security/encryption/default-encryption

> Throughout the remainder of this documentation, the term Key Encryption Key
> can be used interchangeably with Root Key.

## Why use Praetorian?

Praetorian is intended to be very easy to setup and great for startups looking
to get moving quickly. If you're storing sensitive information in a database and
want a simplified approach to using envelope encryption, you should use
Praetorian.

## Deployment

Before you start, you'll need to create a valid Praetorian configuration. Run
the following command from your local terminal to generate a JSON configuration
which contains a single root key which will be set as the active root key.

```text
echo $(printf '{"activeKeyId": "1", "rootKeys": {"1": "%s"}}' "$(openssl rand -base64 32)")
```

Hit the deployment button below and paste the JSON output from the previous
command into the `PRAETORIAN_CONFIG` environment variable. Once the service has
been built and deployed, Praetorian will be running and only available from the
private interface.

> Do not expose your Praetorian service to the public internet!

[![Deploy on Railway](https://railway.com/button.svg)](https://railway.com/template/pr7YF_?referralCode=O1su6l)

When you deploy applications that interact with Praetorian, you will need to use `http://praetorian` as the base URL.

## Usage

Before integrating Praetorian into your application, it's best to start with a locally running instance so you can explore and familiarise yourself with the envelope encryption flow.

Assuming you have Docker installed, start by running the following command to get a Praetorian container running and listening for requests on port `3000`.

```text
docker run --rm \
  --name=praetorian \
  --publish=3000:3000 \
  --env=PRAETORIAN_CONFIG="$(printf '{"activeKeyId": "1", "rootKeys": {"1": "%s"}}' "$(openssl rand -base64 32)")" \
  karlbateman/praetorian
```

Now we have our Praetorian service running locally, we can issue a request to wrap a data encryption key using the following `curl` request.

```text
curl --silent \
  --request POST \
  --data '{"key": "abc123"}' \
  http://localhost:3000/wrap
```

To unwrap this data, simply copy the response from the terminal and send it to the `/unwrap` endpoint.

```text
curl --silent \
  --request POST \
  --data '<replace_with_wrap_response>' \
  http://localhost:3000/unwrap
```

This demonstrates the flow of wrapping and unwrapping data encryption keys.

## Integrating

With Praetorian running in your Railway environment and accessible through a private networking address, the workflow is simple. Within your backend application, you would perform the following operations:

1. Generate a cryptographically secure data encryption key.
2. Use this key to encrypt the sensitive information.
3. Send a `POST` request to `http://praetorian/wrap` with the data encryption key as a JSON request body.
4. Store the response body and the encrypted sensitive information in your DB.

> Every write/update operation to your database should perform this flow to
> ensure maximum security and integrity of your data. Every data encryption key
> should be unique at the row/doc level.

When you want to read the data, you must perform the following steps within your backend application.

1. Read the encrypted data and the wrapped data encryption key.
2. Send a `POST` request to `http://praetorian/unwrap` with the wrapped data encryption key as a JSON request body.
3. Use the unwrapped data encryption key from the response to decrypt the sensitive information.

> Never store an unwrapped data encryption key.
