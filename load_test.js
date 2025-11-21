import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
        { duration: '10s', target: 10 },
        { duration: '20s', target: 10 },
        { duration: '5s', target: 0 },
    ],
    thresholds: {
        http_req_duration: ['p(95)<300'],
        http_req_failed: ['rate<0.001'],
    },
};

const BASE_URL = 'http://localhost:8080';
const TEAM_NAME = `loadtest_team_${randomString(5)}`;

export function setup() {
    console.log(`Creating team: ${TEAM_NAME}`);

    const members = [];
    for (let i = 0; i < 20; i++) {
        members.push({
            user_id: `user_${i}_${randomString(5)}`,
            username: `User ${i}`,
            is_active: true,
        });
    }

    const payload = JSON.stringify({
        team_name: TEAM_NAME,
        members: members,
    });

    const params = { headers: { 'Content-Type': 'application/json' } };
    const res = http.post(`${BASE_URL}/team/add`, payload, params);

    if (res.status !== 201 && res.status !== 200) {
        throw new Error(`Failed to setup team: ${res.status} ${res.body}`);
    }

    return { authorId: members[0].user_id };
}

export default function (data) {
    const prId = `pr_${randomString(10)}`;

    const payload = JSON.stringify({
        pull_request_id: prId,
        pull_request_name: `Feature ${randomString(5)}`,
        author_id: data.authorId,
    });

    const params = { headers: { 'Content-Type': 'application/json' } };

    const res = http.post(`${BASE_URL}/pullRequest/create`, payload, params);

    check(res, {
        'status is 201': (r) => r.status === 201,
        'reviewers assigned': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body.pr.assigned_reviewers && body.pr.assigned_reviewers.length > 0;
            } catch(e) { return false; }
        }
    });

    sleep(0.1);
}