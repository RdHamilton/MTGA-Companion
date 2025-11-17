import { useState, useEffect } from 'react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import { GetActiveEvents, GetEventWinDistribution } from '../../wailsjs/go/main/App';
import { models, storage } from '../../wailsjs/go/models';
import LoadingSpinner from '../components/LoadingSpinner';
import EmptyState from '../components/EmptyState';
import ErrorState from '../components/ErrorState';
import './Events.css';

const Events = () => {
  const [activeEvents, setActiveEvents] = useState<models.DraftEvent[]>([]);
  const [winDistribution, setWinDistribution] = useState<storage.EventWinDistribution[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadEventData();
  }, []);

  // Listen for real-time updates
  useEffect(() => {
    const unsubscribeStats = EventsOn('stats:updated', () => {
      loadEventData();
    });

    return () => {
      if (unsubscribeStats) unsubscribeStats();
    };
  }, []);

  const loadEventData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [active, distribution] = await Promise.all([
        GetActiveEvents(),
        GetEventWinDistribution(),
      ]);

      setActiveEvents(active || []);
      setWinDistribution(distribution || []);
    } catch (err) {
      console.error('Failed to load event data:', err);
      setError(err instanceof Error ? err.message : 'Failed to load event data');
    } finally {
      setLoading(false);
    }
  };

  const formatEventDate = (dateStr: string): string => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getRecordColor = (record: string): string => {
    const [wins] = record.split('-').map(Number);
    if (wins >= 7) return 'record-excellent'; // 7+ wins
    if (wins >= 5) return 'record-good'; // 5-6 wins
    if (wins >= 3) return 'record-average'; // 3-4 wins
    return 'record-poor'; // 0-2 wins
  };

  const getTotalEvents = (): number => {
    return winDistribution.reduce((sum, item) => sum + item.count, 0);
  };

  const getMaxCount = (): number => {
    return Math.max(...winDistribution.map(item => item.count), 1);
  };

  if (loading) {
    return (
      <div className="events-page">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="events-page">
        <ErrorState
          message="Failed to load event data"
          error={error}
          helpText="Check that the database is accessible and try refreshing the page."
        />
      </div>
    );
  }

  return (
    <div className="events-page">
      <h1>Limited Events</h1>

      {/* Active Events Section */}
      <section className="events-section">
        <h2>Active Events</h2>
        {activeEvents.length === 0 ? (
          <EmptyState
            icon="🎫"
            title="No active events"
            message="You don't have any ongoing limited events at the moment."
            helpText="Join a draft or sealed event in MTG Arena to start tracking your event performance."
          />
        ) : (
          <div className="active-events-grid">
            {activeEvents.map((event) => (
              <div key={event.ID} className="active-event-card">
                <div className="event-header">
                  <h3>{event.EventName}</h3>
                  <span className="event-set">{event.SetCode}</span>
                </div>
                <div className="event-record">
                  <span className="record-label">Current Record:</span>
                  <span className={`record-value ${getRecordColor(`${event.Wins}-${event.Losses}`)}`}>
                    {event.Wins}-{event.Losses}
                  </span>
                </div>
                <div className="event-meta">
                  <span>Started: {formatEventDate(event.StartTime as string)}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Win Distribution Section */}
      <section className="events-section">
        <h2>Event Win Distribution</h2>
        {winDistribution.length === 0 ? (
          <EmptyState
            icon="📊"
            title="No event history"
            message="Complete some limited events to see your win distribution statistics."
            helpText="Your event records (0-3, 7-0, etc.) will be displayed here once you finish some drafts or sealed events."
          />
        ) : (
          <>
            <div className="distribution-summary">
              <span>Total Events: {getTotalEvents()}</span>
            </div>
            <div className="distribution-chart">
              {winDistribution.map((item) => {
                const percentage = (item.count / getTotalEvents()) * 100;
                const barWidth = (item.count / getMaxCount()) * 100;

                return (
                  <div key={item.record} className="distribution-row">
                    <span className={`record-label ${getRecordColor(item.record)}`}>
                      {item.record}
                    </span>
                    <div className="bar-container">
                      <div
                        className={`bar ${getRecordColor(item.record)}`}
                        style={{ width: `${barWidth}%` }}
                      >
                        <span className="bar-count">{item.count}</span>
                      </div>
                    </div>
                    <span className="percentage">{percentage.toFixed(1)}%</span>
                  </div>
                );
              })}
            </div>
          </>
        )}
      </section>
    </div>
  );
};

export default Events;
