import { Stack, Text, Title } from "@mantine/core";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/$orgId/")({
  component: TransactionsIndex,
});

function TransactionsIndex() {
  return (
    <Stack>
      <Title order={2}>取引</Title>
      <Text c="dimmed">(未実装)</Text>
    </Stack>
  );
}
