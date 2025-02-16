import { check, fail } from "k6";
import http from "k6/http";
import { Options } from "k6/options";

export const options: Options = {
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(99)<10"],
  },
};

export default function () {
  const res = http.get("http://localhost:5000/health");

  check(res, {
    "health ok": (r) => r.status === 200 || fail("Health check failed"),
  });
}
