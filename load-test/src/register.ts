import { check, fail } from "k6";
import exec from "k6/execution";
import http from "k6/http";
import { Counter } from "k6/metrics";
import { Options } from "k6/options";

export const options: Options = {
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(99)<1000"],
  },
};

const usersCreated = new Counter("users_created");
const usersLoggedIn = new Counter("users_logged_in");

export default function () {
  const uid = exec.scenario.iterationInTest;
  const body = {
    name: `User ${uid}`,
    email: `testuser${uid}@example.com`,
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
      fail(`Registration failed for ${body.name} ${body.email}`),
  });

  if (registration.status === 200) {
    usersLoggedIn.add(1);
  }

  if (registration.status === 201) {
    usersCreated.add(1);
  }
}
