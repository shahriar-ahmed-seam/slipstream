import { MetricsDashboard } from "../../components/MetricsDashboard";

export const dynamic = "force-dynamic";

export default function MetricsPage() {
  return (
    <main className="space-y-6">
      <MetricsDashboard />
    </main>
  );
}