import styles from "../../assets/styles/StatisticsAnalysis.module.scss";

type StatisticsMessageSectionProps = {
  title: string;
  pageMessage: string;
  pageError: string | null;
  error: string | null;
  showSuccess: boolean;
};

const StatisticsMessageSection = ({
  title,
  pageMessage,
  pageError,
  error,
  showSuccess,
}: StatisticsMessageSectionProps) => {
  return (
    <section className={styles.panel}>
      <h2 className={styles.sectionTitle}>{title}</h2>
      <p className={styles.summary}>{pageMessage}</p>
      {error && <div className={styles.errorBox}>{error}</div>}
      {pageError && <div className={styles.errorBox}>{pageError}</div>}
      {!error && !pageError && showSuccess && <div className={styles.successBox}>{pageMessage}</div>}
    </section>
  );
};

export default StatisticsMessageSection;
