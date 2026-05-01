import { apiClient } from "@/api/client";
import { mutationOptions } from "@tanstack/react-query";
import * as v from "valibot";

export const NewTransactionSchema = v.object({
  category_id: v.pipe(
    v.string(),
    v.nonEmpty("カテゴリを選択してください"),
  ),
  amount: v.pipe(
    v.number("金額を入力してください"),
    v.integer("整数で入力してください"),
    v.minValue(1, "1 円以上で入力してください"),
  ),
  date: v.pipe(
    v.string(),
    v.regex(/^\d{4}-\d{2}-\d{2}$/, "日付を選択してください"),
  ),
  memo: v.pipe(v.string(), v.maxLength(500, "500 文字以内で入力してください")),
});

export type NewTransactionInput = v.InferOutput<typeof NewTransactionSchema>;

export const createTransactionMutation = (orgId: string) =>
  mutationOptions({
    mutationKey: ["orgs", orgId, "transactions", "create"],
    mutationFn: async (input: NewTransactionInput) => {
      const { data, error } = await apiClient.POST(
        "/orgs/{org_id}/transactions",
        {
          params: { path: { org_id: orgId } },
          body: input,
        },
      );
      if (error) throw error;
      return data;
    },
  });
