import { categoriesQuery } from "@/features/category/queries";
import {
  NewTransactionSchema,
  createTransactionMutation,
} from "@/features/transaction/mutations";
import { transactionsQuery } from "@/features/transaction/queries";
import {
  Alert,
  Button,
  ColorSwatch,
  Drawer,
  Group,
  NumberInput,
  Select,
  Stack,
  Textarea,
  TextInput,
} from "@mantine/core";
import { schemaResolver, useForm } from "@mantine/form";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";
import { Link, createFileRoute, useNavigate } from "@tanstack/react-router";
import { format } from "date-fns";

export const Route = createFileRoute("/_authed/orgs/$orgId/transactions/new")({
  component: NewTransactionDrawer,
});

function NewTransactionDrawer() {
  const { orgId } = Route.useParams();
  const { month } = Route.useSearch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { data: categories } = useSuspenseQuery(categoriesQuery(orgId));

  const noCategories = categories.items.length === 0;

  const close = () =>
    navigate({
      to: "/orgs/$orgId/transactions",
      params: { orgId },
      search: { month },
    });

  const mutation = useMutation({
    ...createTransactionMutation(orgId),
    onSuccess: async (created) => {
      const createdMonth = created.date.slice(0, 7);
      // Invalidate the displayed month and the created tx's month if they differ.
      await queryClient.invalidateQueries({
        queryKey: transactionsQuery(orgId, month).queryKey,
      });
      if (createdMonth !== month) {
        await queryClient.invalidateQueries({
          queryKey: transactionsQuery(orgId, createdMonth).queryKey,
        });
      }
      close();
    },
  });

  const form = useForm({
    initialValues: {
      category_id: categories.items[0]?.id ?? "",
      amount: 0,
      date: format(new Date(), "yyyy-MM-dd"),
      memo: "",
    },
    validate: schemaResolver(NewTransactionSchema, { sync: true }),
  });

  const submitError = mutation.isError
    ? ((mutation.error as { message?: string })?.message ??
      "作成に失敗しました")
    : null;

  const selectedCat = categories.items.find(
    (c) => c.id === form.values.category_id,
  );

  return (
    <Drawer
      opened
      onClose={close}
      position="bottom"
      size="auto"
      title="取引を追加"
    >
      {noCategories ? (
        <Stack>
          <Alert color="yellow">
            先にカテゴリを作成してください。
          </Alert>
          <Group justify="flex-end">
            <Button variant="subtle" onClick={close} type="button">
              閉じる
            </Button>
            <Link
              to="/orgs/$orgId/categories/new"
              params={{ orgId }}
              style={{ textDecoration: "none" }}
            >
              <Button component="span">カテゴリを作成</Button>
            </Link>
          </Group>
        </Stack>
      ) : (
        <form onSubmit={form.onSubmit((values) => mutation.mutate(values))}>
          <Stack>
            <TextInput
              type="date"
              label="日付"
              {...form.getInputProps("date")}
              required
            />
            <Select
              label="カテゴリ"
              {...form.getInputProps("category_id")}
              data={categories.items.map((c) => ({
                value: c.id,
                label: c.name,
              }))}
              leftSection={
                selectedCat ? (
                  <ColorSwatch color={selectedCat.color} size={14} />
                ) : null
              }
              searchable
              required
            />
            <NumberInput
              label="金額"
              {...form.getInputProps("amount")}
              min={1}
              thousandSeparator=","
              hideControls
              required
            />
            <Textarea
              label="メモ"
              {...form.getInputProps("memo")}
              maxLength={500}
              autosize
              minRows={2}
            />
            {submitError && <Alert color="red">{submitError}</Alert>}
            <Group justify="flex-end">
              <Button variant="subtle" onClick={close} type="button">
                キャンセル
              </Button>
              <Button type="submit" loading={mutation.isPending}>
                作成
              </Button>
            </Group>
          </Stack>
        </form>
      )}
    </Drawer>
  );
}
