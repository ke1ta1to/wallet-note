import { Stack, Text, Title } from "@mantine/core";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/summary")({
  component: SummaryPage,
});

function SummaryPage() {
  return (
    <Stack>
      <Title order={2}>集計</Title>
      <Text c="dimmed">(未実装)</Text>
    </Stack>
  );
}
