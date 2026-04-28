import {
  createOrgMutation,
  NewOrgSchema,
} from "@/features/organization/mutations";
import { meOrgsQuery } from "@/features/organization/queries";
import {
  Alert,
  Anchor,
  Button,
  Container,
  Group,
  Stack,
  TextInput,
  Title,
} from "@mantine/core";
import { schemaResolver, useForm } from "@mantine/form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Link, createFileRoute, useNavigate } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed/orgs/new")({
  component: NewOrg,
});

function NewOrg() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const mutation = useMutation({
    ...createOrgMutation,
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: meOrgsQuery.queryKey });
      navigate({ to: "/orgs" });
    },
  });

  const form = useForm({
    initialValues: { name: "" },
    validate: schemaResolver(NewOrgSchema, { sync: true }),
  });

  const submitError = mutation.isError
    ? ((mutation.error as { message?: string })?.message ??
      "作成に失敗しました")
    : null;

  return (
    <Container size="sm" py="xl">
      <Anchor
        component={Link}
        to="/orgs"
        size="sm"
        mb="md"
        display="inline-block"
      >
        ← Organizations
      </Anchor>
      <Title order={1} mb="md">
        新規組織を作成
      </Title>
      <form onSubmit={form.onSubmit((values) => mutation.mutate(values))}>
        <Stack>
          <TextInput
            label="組織名"
            {...form.getInputProps("name")}
            maxLength={100}
            required
            autoFocus
          />
          {submitError && <Alert color="red">{submitError}</Alert>}
          <Group justify="flex-end">
            <Button type="submit" loading={mutation.isPending}>
              作成
            </Button>
          </Group>
        </Stack>
      </form>
    </Container>
  );
}
