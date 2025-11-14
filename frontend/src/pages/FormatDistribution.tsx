const FormatDistribution = () => {
  return (
    <div className="page-container">
      <h1 className="page-title">Format Distribution</h1>
      <div className="no-data">
        This view is not yet implemented.
        <br /><br />
        Follow the pattern from Win Rate Trend:
        <br />
        1. Import GetStatsByFormat from wailsjs
        <br />
        2. Use Recharts PieChart or BarChart
        <br />
        3. Add filters for date range
        <br />
        4. Show match count and win rate per format
      </div>
    </div>
  );
};

export default FormatDistribution;
