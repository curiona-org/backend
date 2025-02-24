import { check, fail, randomSeed } from "k6";
import exec from "k6/execution";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { Options } from "k6/options";
export const options: Options = {
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(99)<1500"],
  },
  scenarios: {
    ramping_arrival_rate: {
      executor: "ramping-arrival-rate",
      startRate: 1000, // 1000 RPS
      timeUnit: "1s", // 1000 iterations per second, i.e. 1000 RPS
      stages: [
        { target: 1000, duration: "10s" },
        { target: 10000, duration: "20s" },
        { target: 1000, duration: "10s" },
      ],
      preAllocatedVUs: 100, // how large the initial pool of VUs would be
      maxVUs: 300,
    },
  },
};

const accountsRegisteredCounter = new Counter("accounts_registered");

export default function () {
  randomSeed(Date.now());
  const uid = `${Math.random().toString()}${exec.scenario.iterationInTest}`;
  const body = {
    name: `${uid}`,
    email: `${uid}@example.com`,
    password: "123123",
  };

  const registration = http.post(
    "http://localhost:5000/auth",
    JSON.stringify(body),
    {
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
    }
  );

  check(registration, {
    "user registration successful": (r) =>
      [200, 201].includes(r.status) ||
      fail(`${r.status} Registration failed for ${body.name} ${body.email}: ${r.body}`),
  });

  if (registration.status === 201) {
    accountsRegisteredCounter.add(1);
  }
}
