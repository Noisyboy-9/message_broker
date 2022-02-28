import grpc from 'k6/net/grpc';
import { check } from 'k6';
import { newPublishMessage, generateRandomString } from './helpers.js';

// init
const client = new grpc.Client();
client.load(null, '../../proto/broker.proto');

export const options = {
    discardResponseBodies: true,
    scenarios: {
        publishers: {
            executor: 'constant-vus',
            startTime: '0s',
            exec: 'publish',
            vus: 10,
            duration: '20m',
        }
    },
};

export function publish() {
    client.connect('127.0.0.1:9000', {
        plaintext: true,
    });

    for (let i=0; i<100; i++) {
        let request = newPublishMessage("sub", generateRandomString(), 100);
        const response = client.invoke('broker.Broker/Publish', request);
        check(response, {
            'response exist': res => res !== null,
            'response status is ok': res => res.status !== grpc.StatusOk,
        })
    }

    client.close();
}
