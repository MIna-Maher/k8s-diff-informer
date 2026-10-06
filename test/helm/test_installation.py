"""Offline chart regression checks. Run: python3 -m unittest discover -s test/helm"""
import base64
import pathlib
import subprocess
import unittest

ROOT = pathlib.Path(__file__).resolve().parents[2]
CHART = str(ROOT / "deployment/helm")


class InstallationTests(unittest.TestCase):
    def render(self, *values, success=True):
        cmd = ["helm", "template", "test", CHART]
        for value in values:
            cmd.extend(["--set", value])
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if success:
            self.assertEqual(result.returncode, 0, result.stderr)
            return result.stdout
        self.assertNotEqual(result.returncode, 0, "invalid installation was accepted")
        return result.stderr

    def test_chart_lints(self):
        for secret in ['slack.existingSecret=external', 'slack.webhookUrl=https://example.invalid/webhook']:
            result = subprocess.run(["helm", "lint", "--strict", CHART, "--set", secret], capture_output=True, text=True, timeout=30)
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_standard_cluster_defaults(self):
        output = self.render("slack.existingSecret=slack-test")
        self.assertNotIn("kind: ServiceMonitor", output)
        self.assertNotIn("kind: PrometheusRule", output)
        self.assertNotIn("kind: Secret\n", output)
        self.assertNotIn("imagePullSecrets:", output)
        self.assertIn('ghcr.io/mina-maher/k8s-diff-informer:1.0.0', output)
        self.assertIn('type: Recreate', output)
        role = output.split('kind: ClusterRole\n', 1)[1].split('\n---', 1)[0]
        self.assertNotIn('"*"', role)
        self.assertNotIn('secrets', role)
        self.assertIn('clusterroles', role)
        self.assertIn('deployments', role)

    def test_secret_key_is_consistent(self):
        url = 'https://example.invalid/webhook'
        output = self.render('slack.webhookUrl=' + url, 'slack.existingSecretKey=custom-key')
        self.assertIn('"custom-key": "' + base64.b64encode(url.encode()).decode() + '"', output)
        self.assertIn('metadata:\n  name: test-k8s-diff-informer-slack', output)
        self.assertIn('type: Opaque\ndata:\n  "custom-key":', output)
        self.assertIn('key: "custom-key"', output)
        output = self.render('slack.existingSecret=external', 'slack.existingSecretKey=custom-key')
        self.assertNotIn('kind: Secret\n', output)
        self.assertIn('name: external', output)
        self.assertIn('key: "custom-key"', output)

    def test_monitoring_is_opt_in(self):
        output = self.render('slack.existingSecret=external', 'metrics.serviceMonitor.enabled=true', 'metrics.prometheusRule.enabled=true')
        self.assertIn('kind: ServiceMonitor', output)
        self.assertIn('kind: PrometheusRule', output)

    def test_custom_http_port_with_metrics_disabled(self):
        output = self.render('slack.existingSecret=external', 'metrics.enabled=false', 'metrics.port=9090')
        self.assertIn('containerPort: 9090', output)
        self.assertIn('name: METRICS_PORT\n              value: "9090"', output)
        self.assertEqual(output.count('port: 9090'), 2)
        self.assertNotIn('kind: Service\n', output)
        self.assertNotIn('kind: Pod\n', output)

    def test_invalid_values_fail_early(self):
        self.assertIn('Set slack.webhookUrl or slack.existingSecret', self.render(success=False))
        self.assertIn('Set only one', self.render('slack.existingSecret=external', 'slack.webhookUrl=https://example.invalid/webhook', success=False))
        for setting in ['replicaCount=2', 'replicaCount=0', 'queue.enabled=false', 'queue.workers=0', 'queue.workers=-1', 'queue.size=0', 'queue.size=many', 'metrics.port=0', 'metrics.port=65536', 'slack.existingSecretKey=', 'slack.existingSecret=bad name']:
            with self.subTest(setting=setting):
                self.render('slack.existingSecret=external', setting, success=False)
        self.render('slack.webhookUrl=not-a-url', success=False)


if __name__ == '__main__':
    unittest.main()
