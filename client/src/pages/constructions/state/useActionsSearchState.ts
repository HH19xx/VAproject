import { useState } from "react";
import type { ActionLog } from "../../../hooks/useActions";
import type { SortKey, SortOrder } from "../actionsSearch/actionsSearchTypes";

const useActionsSearchState = () => {
  const [sort, setSort] = useState<SortKey>("occurred_at");
  const [order, setOrder] = useState<SortOrder>("desc");
  const [from, setFrom] = useState("");
  const [to, setTo] = useState("");
  const [isDateRangeManual, setIsDateRangeManual] = useState(false);
  const [filterTagIDs, setFilterTagIDs] = useState<number[]>([]);
  const [danbooruQuery, setDanbooruQuery] = useState("");
  const [searchedActionLogs, setSearchedActionLogs] = useState<ActionLog[] | null>(null);
  const [queryError, setQueryError] = useState<string | null>(null);
  const [saveMessage, setSaveMessage] = useState<string | null>(null);

  return {
    sort,
    setSort,
    order,
    setOrder,
    from,
    setFrom,
    to,
    setTo,
    isDateRangeManual,
    setIsDateRangeManual,
    filterTagIDs,
    setFilterTagIDs,
    danbooruQuery,
    setDanbooruQuery,
    searchedActionLogs,
    setSearchedActionLogs,
    queryError,
    setQueryError,
    saveMessage,
    setSaveMessage,
  };
};

export default useActionsSearchState;
