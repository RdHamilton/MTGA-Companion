const DeckPerformance = () => {
  return (
    <div className="page-container">
      <h1 className="page-title">Deck Performance</h1>
      <div className="no-data">
        This view is not yet implemented.
        <br /><br />
        Follow the pattern from Win Rate Trend:
        <br />
        1. Import GetStatsByDeck from wailsjs
        <br />
        2. Use Recharts BarChart for horizontal bars
        <br />
        3. Add filters for date range and format
        <br />
        4. Display deck names with win rates
      </div>
    </div>
  );
};

export default DeckPerformance;
