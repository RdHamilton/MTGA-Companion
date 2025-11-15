import { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetActiveQuests, GetQuestHistory, GetQuestStats } from '../../wailsjs/go/main/App';
import { models } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import Tooltip from '../components/Tooltip';
import './Quests.css';

const Quests = () => {
  const [activeQuests, setActiveQuests] = useState<models.Quest[]>([]);
  const [questHistory, setQuestHistory] = useState<models.Quest[]>([]);
  const [questStats, setQuestStats] = useState<models.QuestStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters for history
  const [dateRange, setDateRange] = useState('30days');
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');
  const [historyLimit] = useState(50);

  // Pagination for history
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);

  useEffect(() => {
    loadQuestData();
  }, [dateRange, customStartDate, customEndDate]);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribe = EventsOn('stats:updated', () => {
      console.log('Stats updated event received - reloading quest data');
      loadQuestData();
    });

    return () => {
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, [dateRange, customStartDate, customEndDate]);

  const loadQuestData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Build date range for history and stats
      let startDate = '';
      let endDate = '';

      if (dateRange === 'custom') {
        startDate = customStartDate;
        endDate = customEndDate;
      } else if (dateRange !== 'all') {
        const now = new Date();
        const start = new Date();

        switch (dateRange) {
          case '7days':
            start.setDate(now.getDate() - 7);
            break;
          case '30days':
            start.setDate(now.getDate() - 30);
            break;
          case '90days':
            start.setDate(now.getDate() - 90);
            break;
        }

        startDate = start.toISOString().split('T')[0];
        endDate = now.toISOString().split('T')[0];
      }

      // Load all quest data in parallel
      const [active, history, stats] = await Promise.all([
        GetActiveQuests(),
        GetQuestHistory(startDate, endDate, historyLimit),
        GetQuestStats(startDate, endDate),
      ]);

      setActiveQuests(active || []);
      setQuestHistory(history || []);
      setQuestStats(stats);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load quest data');
      console.error('Error loading quest data:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (timestamp: any) => {
    return new Date(timestamp).toLocaleDateString();
  };

  const calculateProgress = (quest: models.Quest): number => {
    if (quest.goal === 0) return 0;
    const progress = (quest.ending_progress / quest.goal) * 100;
    return Math.min(progress, 100);
  };

  const formatCompletionTime = (assignedAt: any, completedAt: any): string => {
    if (!completedAt) return 'N/A';

    const assigned = new Date(assignedAt).getTime();
    const completed = new Date(completedAt).getTime();
    const durationMs = completed - assigned;

    const hours = Math.floor(durationMs / (1000 * 60 * 60));
    const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  // Paginate history
  const totalPages = Math.ceil(questHistory.length / pageSize);
  const paginatedHistory = questHistory.slice((page - 1) * pageSize, page * pageSize);

  const getTodayDateString = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  const getMinEndDate = () => {
    return customStartDate || undefined;
  };

  return (
    <div className="page-container">
      {/* Header */}
      <div className="quests-header">
        <h1 className="page-title">Daily Quests</h1>

        {/* Quest Statistics Summary */}
        {!loading && !error && questStats && (
          <div className="quest-stats-summary">
            <div className="stat-card">
              <div className="stat-label">Active Quests</div>
              <div className="stat-value">{questStats.active_quests}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Completion Rate</div>
              <div className="stat-value">{questStats.completion_rate.toFixed(1)}%</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Total Gold Earned</div>
              <div className="stat-value">{questStats.total_gold_earned.toLocaleString()}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Avg Completion Time</div>
              <div className="stat-value">
                {questStats.average_completion_ms > 0
                  ? `${(questStats.average_completion_ms / (1000 * 60 * 60)).toFixed(1)}h`
                  : 'N/A'}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Rerolls Used</div>
              <div className="stat-value">{questStats.reroll_count}</div>
            </div>
          </div>
        )}
      </div>

      {/* Loading/Error States */}
      {loading && <LoadingSpinner message="Loading quest data..." />}
      {error && <div className="error">{error}</div>}

      {!loading && !error && (
        <>
          {/* Active Quests Section */}
          <div className="quests-section">
            <h2 className="section-title">Active Quests</h2>
            {activeQuests.length === 0 ? (
              <div className="no-data">No active quests</div>
            ) : (
              <div className="active-quests-grid">
                {activeQuests.map((quest) => {
                  const progress = calculateProgress(quest);
                  return (
                    <div key={quest.id} className="quest-card">
                      <div className="quest-card-header">
                        <div className="quest-type">{quest.quest_type || 'Daily Quest'}</div>
                        {quest.can_swap && (
                          <Tooltip content="This quest can be rerolled">
                            <span className="reroll-badge">Rerollable</span>
                          </Tooltip>
                        )}
                      </div>
                      <div className="quest-card-body">
                        <div className="quest-progress-text">
                          {quest.ending_progress} / {quest.goal}
                        </div>
                        <div className="quest-progress-bar">
                          <div
                            className="quest-progress-fill"
                            style={{ width: `${progress}%` }}
                          />
                        </div>
                        <div className="quest-progress-percent">{progress.toFixed(0)}%</div>
                      </div>
                      <div className="quest-card-footer">
                        <div className="quest-assigned">
                          Assigned: {formatDate(quest.assigned_at)}
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>

          {/* Quest History Section */}
          <div className="quests-section">
            <div className="section-header">
              <h2 className="section-title">Quest History</h2>

              {/* Filters */}
              <div className="filter-row">
                <div className="filter-group">
                  <label className="filter-label">Date Range</label>
                  <select value={dateRange} onChange={(e) => setDateRange(e.target.value)}>
                    <option value="7days">Last 7 Days</option>
                    <option value="30days">Last 30 Days</option>
                    <option value="90days">Last 90 Days</option>
                    <option value="all">All Time</option>
                    <option value="custom">Custom Range</option>
                  </select>
                </div>

                {dateRange === 'custom' && (
                  <>
                    <div className="filter-group">
                      <label className="filter-label">Start Date</label>
                      <input
                        type="date"
                        value={customStartDate}
                        max={getTodayDateString()}
                        onChange={(e) => setCustomStartDate(e.target.value)}
                      />
                    </div>

                    <div className="filter-group">
                      <label className="filter-label">End Date</label>
                      <input
                        type="date"
                        value={customEndDate}
                        min={getMinEndDate()}
                        max={getTodayDateString()}
                        onChange={(e) => setCustomEndDate(e.target.value)}
                      />
                    </div>
                  </>
                )}
              </div>
            </div>

            {questHistory.length === 0 ? (
              <div className="no-data">No quest history found for the selected period</div>
            ) : (
              <>
                <div className="quest-history-table-container">
                  <table>
                    <thead>
                      <tr>
                        <th>
                          <Tooltip content="Quest type or description">
                            <span>Type</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Quest goal and progress">
                            <span>Progress</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Quest completion status">
                            <span>Status</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="When the quest was assigned">
                            <span>Assigned</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="When the quest was completed">
                            <span>Completed</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Time taken to complete">
                            <span>Duration</span>
                          </Tooltip>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {paginatedHistory.map((quest) => {
                        const progress = calculateProgress(quest);
                        return (
                          <tr key={quest.id} className={quest.completed ? 'quest-completed' : 'quest-incomplete'}>
                            <td>{quest.quest_type || 'Daily Quest'}</td>
                            <td>
                              <div className="progress-cell">
                                <span>{quest.ending_progress} / {quest.goal}</span>
                                <div className="mini-progress-bar">
                                  <div
                                    className="mini-progress-fill"
                                    style={{ width: `${progress}%` }}
                                  />
                                </div>
                              </div>
                            </td>
                            <td>
                              <span className={`status-badge ${quest.completed ? 'completed' : 'incomplete'}`}>
                                {quest.completed ? 'COMPLETED' : 'INCOMPLETE'}
                              </span>
                            </td>
                            <td>{formatDate(quest.assigned_at)}</td>
                            <td>{quest.completed_at ? formatDate(quest.completed_at) : '-'}</td>
                            <td>{formatCompletionTime(quest.assigned_at, quest.completed_at)}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                  <div className="quest-history-footer">
                    <div className="pagination">
                      <button
                        onClick={() => setPage(1)}
                        disabled={page === 1}
                        className="pagination-btn"
                      >
                        First
                      </button>
                      <button
                        onClick={() => setPage(page - 1)}
                        disabled={page === 1}
                        className="pagination-btn"
                      >
                        Previous
                      </button>
                      <span className="pagination-info">
                        Page {page} of {totalPages}
                      </span>
                      <button
                        onClick={() => setPage(page + 1)}
                        disabled={page === totalPages}
                        className="pagination-btn"
                      >
                        Next
                      </button>
                      <button
                        onClick={() => setPage(totalPages)}
                        disabled={page === totalPages}
                        className="pagination-btn"
                      >
                        Last
                      </button>
                    </div>
                  </div>
                )}
              </>
            )}
          </div>
        </>
      )}
    </div>
  );
};

export default Quests;
