import { useEffect, useState } from "react";
import { systemApi } from "../api/system";
import type { DictionaryOption, DictionaryOptions } from "../models/system";

let optionsRequest: Promise<DictionaryOptions> | undefined;

function loadDictionaryOptions() {
  optionsRequest ??= systemApi.dictionaryOptions().catch((error: unknown) => {
    optionsRequest = undefined;
    throw error;
  });
  return optionsRequest;
}

export function useDictionaryOptions(type: string) {
  const [options, setOptions] = useState<DictionaryOption[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void loadDictionaryOptions()
      .then((allOptions) => {
        if (active) setOptions(allOptions[type] ?? []);
      })
      .catch((cause: unknown) => {
        if (active) setError(cause instanceof Error ? cause.message : "字典加载失败");
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [type]);

  return { options, loading, error };
}

export function dictionaryName(options: DictionaryOption[], code: string | number | undefined) {
  return options.find((option) => option.code === String(code))?.name ?? String(code ?? "");
}
