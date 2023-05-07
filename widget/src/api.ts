import { useEffect, useState } from 'react';

export const call = async <TRequest = null, TResponse = null>(
    method: string,
    path: string,
    body?: TRequest
) => {
    const headers: Record<string, string> = {
        'Content-Type': 'application/json; charset=UTF-8'
    };
    const opts: RequestInit = {
        mode: 'same-origin',
        method: method.toUpperCase()
    };
    if (method !== 'GET' && body) {
        opts['body'] = JSON.stringify(body);
    }
    opts.headers = headers;
    const url = new URL(
        `http://localhost:8900/${path}`,
        `${location.protocol}//${location.host}`
    );

    const resp = await fetch(url, opts);
    if (!resp.ok) {
        throw await resp.json();
    }
    const json: TResponse = await resp.json();
    return json;
};

export type steamIdFormat = 'steam64' | 'steam3' | 'steam32' | 'steam';

enum ProfileVisibility {
    ProfileVisibilityPrivate = 1,
    ProfileVisibilityFriendsOnly = 2,
    ProfileVisibilityPublic = 3
}
export enum Team {
    SPEC,
    UNASSIGNED,
    BLU,
    RED
}

export interface Match {
    origin: string;
    attributes: string[];
    matcher_type: string;
}

export interface Player {
    steam_id: bigint;
    name: string;
    created_on: Date;
    updated_on: Date;
    team: Team;
    profile_updated_on: Date;
    kills_on: number;
    rage_quits: number;
    deaths_by: number;
    notes: string;
    whitelisted: boolean;
    real_name: string;
    name_previous: string;
    account_created_on: Date;
    visibility: ProfileVisibility;
    avatar_hash: string;
    community_banned: boolean;
    number_of_vac_bans: number;
    last_vac_ban_on: Date | null;
    number_of_game_bans: number;
    economy_ban: boolean;
    connected: number;
    user_id: number;
    ping: number;
    kills: number;
    deaths: number;
    our_friend: boolean;
    matches: Match[];
}

export const formatSeconds = (seconds: number): string => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.round(seconds % 60);
    return [h, m > 9 ? m : h ? '0' + m : m || '0', s > 9 ? s : '0' + s]
        .filter(Boolean)
        .join(':');
};

interface List {
    list_type: string;
    name: string;
    enabled: boolean;
    url: string;
}

export interface Link {
    enabled: boolean;
    name: string;
    url: string;
    id_format: steamIdFormat;
    deleted: boolean;
}

export interface UserSettings {
    steam_id: string;
    steam_dir: string;
    tf2_dir: string;
    auto_launch_game: boolean;
    auto_close_on_game_exit: boolean;
    api_key: string;
    disconnected_timeout: string;
    discord_presence_enabled: boolean;
    kicker_enabled: boolean;
    chat_warnings_enabled: boolean;
    party_warnings_enabled: boolean;
    kick_tags: string[];
    voice_bans_enabled: boolean;
    debug_log_enabled: boolean;
    lists: List[];
    links: Link[];
    rcon_static: boolean;
    gui_enabled: boolean;
    http_enabled: boolean;
    http_listen_addr: string;
    player_expired_timeout: number;
    player_disconnect_timeout: number;
    unique_tags: string[];
}

export const getPlayers = async () => {
    return await call<null, Player[]>('GET', 'players');
};

export const getUserSettings = async () => {
    return await call<null, UserSettings>('GET', 'settings');
};

export const useUserSettings = () => {
    const [settings, setSettings] = useState<UserSettings>();
    const [error, setError] = useState<unknown>(null);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        setLoading(true);
        getUserSettings()
            .then((resp) => resp)
            .then(setSettings)
            .catch(setError)
            .finally(() => setLoading(false));
    }, []);
    return { settings, error, loading };
};

export const usePlayers = () => {
    const [players, setPlayers] = useState<Player[]>([]);
    useEffect(() => {
        const interval = setInterval(async () => {
            try {
                setPlayers(await getPlayers());
            } catch (e) {
                console.log(e);
            }
        }, 1000);
        return () => {
            clearInterval(interval);
        };
    }, []);
    return players;
};
