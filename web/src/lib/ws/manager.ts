export type WSConnectionState = 'idle' | 'connecting' | 'open' | 'closed';

export class WSManager {
  private state: WSConnectionState = 'idle';

  getState(): WSConnectionState {
    return this.state;
  }
}
