import { check, fail } from "k6";
import http from "k6/http";
import { Options } from "k6/options";

export const options: Options = {
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(99)<10"],
  },
  scenarios: {
    ramping_arrival_rate: {
      executor: 'ramping-arrival-rate',
      startRate: 1000, // 1000 RPS
      timeUnit: '1s', // 1000 iterations per second, i.e. 1000 RPS
      stages: [
        { target: 1000, duration: '10s' },
        { target: 10000, duration: '20s' },
        { target: 1000, duration: '10s' },
      ],
      preAllocatedVUs: 100, // how large the initial pool of VUs would be
      maxVUs: 200,
    }
  }
};

export default function () {
  const res = http.get("http://localhost:5000/health");

  check(res, {
    "health ok": (r) => r.status === 200 || fail("Health check failed"),
  });
}
