import styles from "../../assets/styles/SearchSuggestionRail.module.scss";
import type { SearchTagSuggestion } from "../../pages/constructions/components/actionsSearchSectionHelpers";

type SearchSuggestionRailProps = {
  title: string;
  suggestions: SearchTagSuggestion[];
  emptyMessage: string;
  onSelectSuggestion: (tagName: string) => void;
};

const SearchSuggestionRail = ({
  title,
  suggestions,
  emptyMessage,
  onSelectSuggestion,
}: SearchSuggestionRailProps) => {
  return (
    <div className={styles.suggestionBox}>
      <div className={styles.suggestionTitle}>{title}</div>
      {suggestions.length > 0 ? (
        <div className={styles.suggestionList}>
          {suggestions.map((suggestion) => (
            <button
              key={`${suggestion.source}-${suggestion.id}`}
              type="button"
              className={styles.suggestionChip}
              onClick={() => onSelectSuggestion(suggestion.name)}
            >
              #{suggestion.name}
            </button>
          ))}
        </div>
      ) : (
        <div className={styles.suggestionEmpty}>{emptyMessage}</div>
      )}
    </div>
  );
};

export default SearchSuggestionRail;
