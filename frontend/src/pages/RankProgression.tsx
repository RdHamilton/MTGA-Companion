const RankProgression = () => {
  return (
    <div className="page-container">
      <h1 className="page-title">Rank Progression</h1>
      <div className="no-data">
        This view is not yet implemented.
        <br /><br />
        Follow the pattern from Win Rate Trend:
        <br />
        1. Import GetRankProgressionTimeline from wailsjs
        <br />
        2. Use Recharts LineChart to show rank changes over time
        <br />
        3. Add filters for format (Ladder/Play) and date range
        <br />
        4. Convert rank_class/rank_level to numeric values for Y-axis
      </div>
    </div>
  );
};

export default RankProgression;
