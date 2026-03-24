export interface ParsedQuery {
  andTagIDs: number[];
  anyTagIDs: number[];
  anyTagGroups: number[][];
  excludeTagIDs: number[];
  unknownTokens: string[];
}

export const normalizeTagToken = (token: string): string =>
  token.trim().toLowerCase().replace(/\s+/g, "_");

export const parseDanbooruStyleQuery = (query: string, tagNameToID: Map<string, number>): ParsedQuery => {
  const andTagIDs: number[] = [];
  const anyTagIDs: number[] = [];
  const anyTagGroups: number[][] = [];
  const excludeTagIDs: number[] = [];
  const unknownTokens: string[] = [];

  const seenAnd = new Set<number>();
  const seenAny = new Set<number>();
  const seenExclude = new Set<number>();

  const rawTokens = query
    .split(/\s+/)
    .map((token) => token.trim())
    .filter((token) => token.length > 0);

  if (!rawTokens.includes("|")) {
    for (const rawToken of rawTokens) {
      let mode: "and" | "any" | "exclude" = "and";
      let body = rawToken;

      if (body.startsWith("-")) {
        mode = "exclude";
        body = body.slice(1);
      } else if (body.startsWith("~")) {
        mode = "any";
        body = body.slice(1);
      }

      if (!body) continue;

      const tagID = tagNameToID.get(normalizeTagToken(body));
      if (!tagID) {
        unknownTokens.push(rawToken);
        continue;
      }

      if (mode === "and" && !seenAnd.has(tagID)) {
        seenAnd.add(tagID);
        andTagIDs.push(tagID);
      }
      if (mode === "any" && !seenAny.has(tagID)) {
        seenAny.add(tagID);
        anyTagIDs.push(tagID);
      }
      if (mode === "exclude" && !seenExclude.has(tagID)) {
        seenExclude.add(tagID);
        excludeTagIDs.push(tagID);
      }
    }

    return { andTagIDs, anyTagIDs, anyTagGroups, excludeTagIDs, unknownTokens };
  }

  const groups: string[][] = [];
  let currentGroup: string[] = [];
  for (const token of rawTokens) {
    if (token === "|") {
      if (currentGroup.length > 0) groups.push(currentGroup);
      currentGroup = [];
      continue;
    }
    currentGroup.push(token);
  }
  if (currentGroup.length > 0) groups.push(currentGroup);

  for (const group of groups) {
    const groupAnd: number[] = [];
    const seenGroupAnd = new Set<number>();

    for (const rawToken of group) {
      let body = rawToken;

      if (body.startsWith("-")) {
        body = body.slice(1);
        if (!body) continue;
        const tagID = tagNameToID.get(normalizeTagToken(body));
        if (!tagID) {
          unknownTokens.push(rawToken);
          continue;
        }
        if (!seenExclude.has(tagID)) {
          seenExclude.add(tagID);
          excludeTagIDs.push(tagID);
        }
        continue;
      }

      if (body.startsWith("~")) {
        body = body.slice(1);
        if (!body) continue;
        const tagID = tagNameToID.get(normalizeTagToken(body));
        if (!tagID) {
          unknownTokens.push(rawToken);
          continue;
        }
        if (!seenAny.has(tagID)) {
          seenAny.add(tagID);
          anyTagIDs.push(tagID);
        }
        continue;
      }

      if (body.startsWith("+")) {
        body = body.slice(1);
      }
      if (!body) continue;

      const tagID = tagNameToID.get(normalizeTagToken(body));
      if (!tagID) {
        unknownTokens.push(rawToken);
        continue;
      }
      if (!seenGroupAnd.has(tagID)) {
        seenGroupAnd.add(tagID);
        groupAnd.push(tagID);
      }
    }

    if (groupAnd.length > 0) {
      anyTagGroups.push(groupAnd);
    }
  }

  return { andTagIDs, anyTagIDs, anyTagGroups, excludeTagIDs, unknownTokens };
};
