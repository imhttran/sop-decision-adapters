#!/usr/bin/env python3
"""Unit test for the invariant recomputation in the CLEF-002 probe (no network).

    python3 tools/clef/test_probe_systemone.py
"""

import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import probe_systemone  # noqa: E402

# Values copied from captures/live/choice-run1.json and score.json.
LIVE = {"answers": {
    "risk": {"type": "choice", "choice": "HIGH",
             "probabilities": {"LOW": 0.02424653587413193, "MEDIUM": 0.06590891786973317, "HIGH": 0.9098445462561349},
             "confidence": 0.676513609057984},
    "s": {"type": "score", "score": 0.10430811755852966,
          "probabilities": {"0": 0.9218425424821404, "1": 0.05200679747718923, "2": 0.026150660040670213},
          "confidence": 0.7050259489055778},
    "b": {"type": "noul", "noul": 0.94}}}


class ChecksTest(unittest.TestCase):
    def test_live_answers_satisfy_invariants(self):
        out = probe_systemone.checks(LIVE)
        self.assertNotIn("b", out)  # noul has no probabilities to check
        self.assertTrue(out["risk"]["confidence_matches_entropy_formula"])
        self.assertTrue(out["risk"]["choice_is_argmax"])
        self.assertTrue(out["s"]["score_is_expected_index"])

    def test_detects_violations(self):
        bad = {"answers": {"risk": {"type": "choice", "choice": "LOW", "probabilities": {"LOW": 0.2, "HIGH": 0.8},
                                    "confidence": 0.8}}}
        out = probe_systemone.checks(bad)["risk"]
        self.assertFalse(out["confidence_matches_entropy_formula"])
        self.assertFalse(out["choice_is_argmax"])

    def test_error_body_yields_nothing(self):
        self.assertEqual(probe_systemone.checks({"error": "x"}), {})
        self.assertEqual(probe_systemone.checks("not json"), {})


if __name__ == "__main__":
    unittest.main()
