import encoding from 'k6/encoding';

const generateRandomString = (length = 6) => Math.random().toString(20).substring(2, length)
const newPublishMessage = (subject, body, expirationSeconds) => ({
    subject,
    body: encoding.b64encode(body),
    expirationSeconds
});

export {generateRandomString, newPublishMessage}