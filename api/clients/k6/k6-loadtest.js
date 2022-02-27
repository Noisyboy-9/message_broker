import grpc from 'k6/net/grpc';
import encoding from 'k6/encoding';

const client = new grpc.Client();

client.load(null, '../../proto/broker.proto');

const generateRandomString = function(length=6){
    return Math.random().toString(20).substr(2, length)
}

const Duration = '5m'

export const options = {
    discardResponseBodies: true,
    scenarios: {
        publishers: {
            executor: 'constant-vus',
            startTime: '0s',
            exec: 'publish',
            vus: 10,
            duration: Duration,
        }
    },
};
function newPublishMessage(subject, body, expirationSeconds) {
    return {
        subject,
        body: encoding.b64encode(body),
        expirationSeconds
    }
}
export function publish() {
    client.connect('127.0.0.1:9000', {
        plaintext: true, // Note: Solving TLS connection issues.
    });

    for (let i=0; i<100; i++) {
        const response = client.invoke('broker.Broker/Publish', newPublishMessage("sub", generateRandomString(), 100));
    }

    client.close();
}
