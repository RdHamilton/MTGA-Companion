import { useState, useEffect } from 'react';
import { EventsOn } from '@/services/websocketClient';
import { quests, system } from '@/services/api';
import { models } from '@/types/models';
import LoadingSpinner from '../components/LoadingSpinner';
import Tooltip from '../components/Tooltip';
import EmptyState from '../components/EmptyState';
import ErrorState from '../components/ErrorState';
import './Quests.css';

const Quests = () => {
  const [activeQuests, setActiveQuests] = useState<models.Quest[]>([]);
  const [questHistory, setQuestHistory] = useState<models.Quest[]>([]);
  const [currentAccount, setCurrentAccount] = useState<models.Account | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Filters for history
  const [dateRange, setDateRange] = useState('30days');
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');
  const [historyLimit] = useState(50);

  // Table filters
  const [statusFilter, setStatusFilter] = useState<'all' | 'completed' | 'incomplete' | 'rerolled'>('all');
  const [typeFilter, setTypeFilter] = useState('');

  // Pagination for history
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);

  useEffect(() => {
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

      // Load quest data sequentially with better error reporting
      try {
        const active = await quests.getActiveQuests();
        setActiveQuests(active || []);
      } catch (activeErr) {
        console.error('Error loading active quests:', activeErr);
        throw new Error(`Failed to load active quests: ${activeErr instanceof Error ? activeErr.message : String(activeErr)}`);
      }

      try {
        console.log('Loading quest history with dates:', startDate, endDate, historyLimit);
        const history = await quests.getQuestHistory(startDate, endDate, historyLimit);
        console.log('Quest history loaded:', history?.length || 0, 'quests');
        setQuestHistory(history || []);
      } catch (historyErr) {
        console.error('Error loading quest history:', historyErr);
        throw new Error(`Failed to load quest history: ${historyErr instanceof Error ? historyErr.message : String(historyErr)}`);
      }

      try {
        const account = await system.getCurrentAccount();
        setCurrentAccount(account);
      } catch (accountErr) {
        console.error('Error loading current account:', accountErr);
        // Don't throw - account data is optional
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load quest data');
      console.error('Error loading quest data:', err);
    } finally {
      setLoading(false);
    }
  };

    loadQuestData();
  }, [dateRange, customStartDate, customEndDate, historyLimit]);

  // Listen for real-time updates
  useEffect(() => {
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

        // Load quest data sequentially with better error reporting
        try {
          const active = await quests.getActiveQuests();
          setActiveQuests(active || []);
        } catch (activeErr) {
          console.error('Error loading active quests:', activeErr);
          throw new Error(`Failed to load active quests: ${activeErr instanceof Error ? activeErr.message : String(activeErr)}`);
        }

        try {
          console.log('Loading quest history with dates:', startDate, endDate, historyLimit);
          const history = await quests.getQuestHistory(startDate, endDate, historyLimit);
          console.log('Quest history loaded:', history?.length || 0, 'quests');
          setQuestHistory(history || []);
        } catch (historyErr) {
          console.error('Error loading quest history:', historyErr);
          throw new Error(`Failed to load quest history: ${historyErr instanceof Error ? historyErr.message : String(historyErr)}`);
        }

        try {
          const account = await system.getCurrentAccount();
          setCurrentAccount(account);
        } catch (accountErr) {
          console.error('Error loading current account:', accountErr);
          // Don't throw - account data is optional
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load quest data');
        console.error('Error loading quest data:', err);
      } finally {
        setLoading(false);
      }
    };

    const unsubscribeStats = EventsOn('stats:updated', () => {
      console.log('Stats updated event received - reloading quest data');
      loadQuestData();
    });

    const unsubscribeQuests = EventsOn('quest:updated', () => {
      console.log('Quest updated event received - reloading quest data');
      loadQuestData();
    });

    return () => {
      if (unsubscribeStats) unsubscribeStats();
      if (unsubscribeQuests) unsubscribeQuests();
    };
  }, [dateRange, customStartDate, customEndDate, historyLimit]);

  const formatDate = (timestamp: unknown) => {
    return new Date(String(timestamp)).toLocaleDateString();
  };

  const calculateProgress = (quest: models.Quest): number => {
    if (quest.goal === 0) return 0;
    const progress = (quest.ending_progress / quest.goal) * 100;
    return Math.min(progress, 100);
  };

  const formatCompletionTime = (assignedAt: unknown, completedAt: unknown): string => {
    if (!completedAt) return 'N/A';

    const assigned = new Date(String(assignedAt)).getTime();
    const completed = new Date(String(completedAt)).getTime();
    const durationMs = completed - assigned;

    const hours = Math.floor(durationMs / (1000 * 60 * 60));
    const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));

    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  const formatQuestType = (questType: string): string => {
    if (!questType) return 'Quest';

    // Remove "Quests/Quest_" prefix
    let formatted = questType.replace(/^Quests\/Quest_/, '');

    // Replace underscores with spaces
    formatted = formatted.replace(/_/g, ' ');

    return formatted || 'Quest';
  };

  // Filter history
  const filteredHistory = questHistory.filter((quest) => {
    // Status filter
    if (statusFilter !== 'all') {
      if (statusFilter === 'completed' && !quest.completed) return false;
      if (statusFilter === 'incomplete' && (quest.completed || quest.rerolled)) return false;
      if (statusFilter === 'rerolled' && !quest.rerolled) return false;
    }

    // Type filter (search in quest type)
    if (typeFilter) {
      const formattedType = formatQuestType(quest.quest_type).toLowerCase();
      if (!formattedType.includes(typeFilter.toLowerCase())) return false;
    }

    return true;
  });

  // Reset page when filters change
  useEffect(() => {
    setPage(1);
  }, [statusFilter, typeFilter]);

  // Paginate filtered history
  const totalPages = Math.ceil(filteredHistory.length / pageSize);
  const paginatedHistory = filteredHistory.slice((page - 1) * pageSize, page * pageSize);

  const getTodayDateString = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  const getMinEndDate = () => {
    return customStartDate || undefined;
  };

  const getDailyWinsColorClass = (wins: number): string => {
    if (wins < 5) return 'low';     // Red
    if (wins < 15) return 'medium'; // Yellow
    return 'high';                   // Green
  };

  return (
    <div className="page-container">
      {/* Header */}
      <div className="quests-header">
        <h1 className="page-title">Daily Quests</h1>

        {/* Mastery Pass Summary */}
        {!loading && !error && currentAccount && (
          <div className="quest-stats-summary">
            <div className="stat-card">
              <div className="stat-label">Mastery Level</div>
              <div className="stat-value">{currentAccount.MasteryLevel}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Pass Type</div>
              <div className="stat-value">{currentAccount.MasteryPass}</div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Progress</div>
              <div className="stat-value">
                {currentAccount.MasteryMax > 0
                  ? `${((currentAccount.MasteryLevel / currentAccount.MasteryMax) * 100).toFixed(1)}%`
                  : 'N/A'}
              </div>
            </div>
            <div className="stat-card">
              <div className="stat-label">Daily Goal</div>
              <div className="stat-value">{currentAccount.DailyWins >= 5 ? '✓' : `${currentAccount.DailyWins}/5`}</div>
            </div>
          </div>
        )}
      </div>

      {/* Loading/Error States */}
      {loading && <LoadingSpinner message="Loading quest data..." />}
      {error && (
        <ErrorState
          message="Failed to load quest data"
          error={error}
          helpText="Make sure detailed logging is enabled in MTGA: Options → View Account → Detailed Logs (Plugin Support)"
        />
      )}

      {!loading && !error && (
        <>
          {/* Daily/Weekly Wins Section */}
          {currentAccount && (
            <div className="quests-section">
              <h2 className="section-title">Win Progress</h2>
              <div className="wins-grid">
                {/* Daily Wins */}
                <div className="daily-wins-card">
                  <div className="daily-wins-header">
                    <span className="daily-wins-title">Daily Wins</span>
                    <span className="daily-wins-progress">{currentAccount.DailyWins} / 15</span>
                  </div>
                  <div className="daily-wins-bar">
                    <div
                      className={`daily-wins-fill ${getDailyWinsColorClass(currentAccount.DailyWins)}`}
                      style={{ width: `${(currentAccount.DailyWins / 15) * 100}%` }}
                    />
                  </div>
                  <div className="daily-wins-footer">
                    <span className="daily-wins-percent">{((currentAccount.DailyWins / 15) * 100).toFixed(0)}% Complete</span>
                    <span className="daily-wins-reward">
                      {currentAccount.DailyWins < 5 ? 'Goal: 5 wins for mastery' : 'Earn up to 1,250 gold'}
                    </span>
                  </div>
                </div>

                {/* Weekly Wins */}
                <div className="daily-wins-card">
                  <div className="daily-wins-header">
                    <span className="daily-wins-title">Weekly Wins</span>
                    <span className="daily-wins-progress">{currentAccount.WeeklyWins} / 15</span>
                  </div>
                  <div className="daily-wins-bar">
                    <div
                      className="daily-wins-fill weekly"
                      style={{ width: `${(currentAccount.WeeklyWins / 15) * 100}%` }}
                    />
                  </div>
                  <div className="daily-wins-footer">
                    <span className="daily-wins-percent">{((currentAccount.WeeklyWins / 15) * 100).toFixed(0)}% Complete</span>
                    <span className="daily-wins-reward">Earn up to 2,250 gold</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Active Quests Section */}
          <div className="quests-section">
            <h2 className="section-title">Active Quests</h2>
            {activeQuests.length === 0 ? (
              <EmptyState
                icon="📋"
                title="No active quests"
                message="You don't have any active daily quests at the moment."
                helpText="Daily quests reset daily. Make sure detailed logging is enabled in MTGA to track new quests."
              />
            ) : (
              <div className="active-quests-grid">
                {activeQuests.map((quest) => {
                  const progress = calculateProgress(quest);
                  return (
                    <div key={quest.id} className="quest-card">
                      <div className="quest-card-header">
                        <div className="quest-type">
                          {formatQuestType(quest.quest_type)}
                          {quest.rewards && quest.rewards.includes('750') && ' (750 Gold)'}
                          {quest.rewards && quest.rewards.includes('500') && ' (500 Gold)'}
                        </div>
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

                <div className="filter-group">
                  <label className="filter-label">Status</label>
                  <select
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value as 'all' | 'completed' | 'incomplete' | 'rerolled')}
                  >
                    <option value="all">All Status</option>
                    <option value="completed">Completed</option>
                    <option value="incomplete">Incomplete</option>
                    <option value="rerolled">Rerolled</option>
                  </select>
                </div>

                <div className="filter-group">
                  <label className="filter-label">Quest Type</label>
                  <input
                    type="text"
                    placeholder="Search quests..."
                    value={typeFilter}
                    onChange={(e) => setTypeFilter(e.target.value)}
                    className="filter-input"
                  />
                </div>
              </div>
            </div>

            {questHistory.length === 0 ? (
              <EmptyState
                icon="📜"
                title="No quest history"
                message="No completed quests found for the selected time period."
                helpText="Try adjusting the date range or complete some quests to see your history here."
              />
            ) : filteredHistory.length === 0 ? (
              <EmptyState
                icon="🔍"
                title="No matching quests"
                message="No quests match your current filters."
                helpText="Try adjusting your status or type filters to see more results."
              />
            ) : (
              <>
                <div className="quest-history-table-container">
                  <table>
                    <thead>
                      <tr>
                        <th>
                          <Tooltip content="Quest type or description" position="bottom">
                            <span>Type</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Quest goal and progress" position="bottom">
                            <span>Progress</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Quest completion status" position="bottom">
                            <span>Status</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="When the quest was assigned" position="bottom">
                            <span>Assigned</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="When the quest was completed" position="bottom">
                            <span>Completed</span>
                          </Tooltip>
                        </th>
                        <th>
                          <Tooltip content="Time taken to complete" position="bottom">
                            <span>Duration</span>
                          </Tooltip>
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {paginatedHistory.map((quest) => {
                        const progress = calculateProgress(quest);
                        const getRowClass = () => {
                          if (quest.rerolled) return 'quest-rerolled';
                          if (quest.completed) return 'quest-completed';
                          return 'quest-incomplete';
                        };
                        const getStatusBadge = () => {
                          if (quest.rerolled) return { class: 'rerolled', text: 'REROLLED' };
                          if (quest.completed) return { class: 'completed', text: 'COMPLETED' };
                          return { class: 'incomplete', text: 'INCOMPLETE' };
                        };
                        const status = getStatusBadge();
                        return (
                          <tr key={quest.id} className={getRowClass()}>
                            <td>{formatQuestType(quest.quest_type)}</td>
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
                              <span className={`status-badge ${status.class}`}>
                                {status.text}
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
