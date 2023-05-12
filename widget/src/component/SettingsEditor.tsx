import React, {
    ChangeEvent,
    useCallback,
    useContext,
    useEffect,
    useState
} from 'react';
import {
    Button,
    Checkbox,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    FormControl,
    FormControlLabel,
    FormGroup,
    FormHelperText,
    InputLabel,
    ListItemText,
    OutlinedInput,
    Select,
    SelectChangeEvent,
    TextField
} from '@mui/material';
import CheckIcon from '@mui/icons-material/Check';
import Dialog from '@mui/material/Dialog';
import Tooltip from '@mui/material/Tooltip';
import CloseIcon from '@mui/icons-material/Close';
import MenuItem from '@mui/material/MenuItem';
import { UserSettings } from '../api';
import _ from 'lodash';
import { SettingsContext } from '../context/settings';
import Grid2 from '@mui/material/Unstable_Grid2';
import SteamID from 'steamid';

type inputValidator = (value: string) => string | null;

interface SettingsTextBoxProps {
    label: string;
    value: string;
    setValue: (value: string) => void;
    tooltip: string;
    secrets?: boolean;
    validator?: inputValidator;
}

export const SettingsTextBox = ({
    label,
    value,
    setValue,
    tooltip,
    secrets,
    validator
}: SettingsTextBoxProps) => {
    const [error, setError] = useState<string | null>(null);
    const handleChange = useCallback(
        (event: ChangeEvent<HTMLTextAreaElement>) => {
            setValue(event.target.value);
            if (validator) {
                setError(validator(event.target.value));
            }
        },
        [setValue, validator]
    );

    return (
        <Tooltip title={tooltip} placement={'top'}>
            <FormControl fullWidth>
                <TextField
                    hiddenLabel
                    type={secrets ? 'password' : 'text'}
                    error={Boolean(error)}
                    id={`settings-textfield-${label}`}
                    label={label}
                    value={value}
                    onChange={handleChange}
                />
                <FormHelperText>{error}</FormHelperText>
            </FormControl>
        </Tooltip>
    );
};

const validatorSteamID = (value: string): string => {
    let err = 'Invalid SteamID';
    try {
        const id = new SteamID(value);
        if (id.isValid()) {
            err = '';
        }
    } catch (_) {
        /* empty */
    }
    return err;
};

const makeValidatorLength = (length: number): inputValidator => {
    return (value: string): string => {
        if (value.length != length) {
            return 'Invalid value';
        }
        return '';
    };
};

export const isStringIp = (value: string): boolean => {
    return /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/.test(
        value
    );
};

const validatorAddress = (value: string): string => {
    const pcs = value.split(':');
    if (pcs.length != 2) {
        return 'Format must match host:port';
    }
    if (pcs[0].toLowerCase() != 'localhost' && !isStringIp(pcs[0])) {
        return 'Invalid address. x.x.x.x or localhost accepted';
    }
    const port = parseInt(pcs[1], 10);
    if (!/^\d+$/.test(pcs[1])) {
        return 'Invalid port, must be positive integer';
    }
    if (port <= 0 || port > 65535) {
        return 'Invalid port, must be in range: 1-65535';
    }
    return '';
};

interface SettingsCheckBoxProps {
    label: string;
    enabled: boolean;
    setEnabled: (checked: boolean) => void;
    tooltip: string;
}

export const SettingsCheckBox = ({
    label,
    enabled,
    setEnabled,
    tooltip
}: SettingsCheckBoxProps) => {
    return (
        <FormGroup>
            <Tooltip title={tooltip} placement="top">
                <FormControlLabel
                    control={
                        <Checkbox
                            checked={enabled}
                            onChange={(_, checked) => {
                                setEnabled(checked);
                            }}
                        />
                    }
                    label={label}
                />
            </Tooltip>
        </FormGroup>
    );
};

interface SettingsMultiSelectProps {
    label: string;
    values: string[];
    setValues: (values: string[]) => void;
    tooltip: string;
}

export const SettingsMultiSelect = ({
    values,
    setValues,
    label,
    tooltip
}: SettingsMultiSelectProps) => {
    const { settings } = useContext(SettingsContext);
    const handleChange = (event: SelectChangeEvent<typeof values>) => {
        const {
            target: { value }
        } = event;
        setValues(typeof value === 'string' ? value.split(',') : value);
    };

    return (
        <Tooltip title={tooltip} placement="top">
            <FormControl fullWidth>
                <InputLabel id={`settings-select-${label}-label`}>
                    {label}
                </InputLabel>
                <Select<string[]>
                    fullWidth
                    labelId={`settings-select-${label}-label`}
                    id={`settings-select-${label}`}
                    multiple
                    value={values}
                    defaultValue={values}
                    onChange={handleChange}
                    input={<OutlinedInput label="Tag" />}
                    renderValue={(selected) => selected.join(', ')}
                >
                    {settings.kick_tags.map((name) => (
                        <MenuItem key={name} value={name}>
                            <Checkbox checked={values.indexOf(name) > -1} />
                            <ListItemText primary={name} />
                        </MenuItem>
                    ))}
                </Select>
            </FormControl>
        </Tooltip>
    );
};

interface SettingsEditorProps {
    open: boolean;
    setOpen: (opeN: boolean) => void;
    origSettings: UserSettings;
}

export const SettingsEditor = ({
    open,
    setOpen,
    origSettings
}: SettingsEditorProps) => {
    const [settings, setSettings] = useState<UserSettings>(
        _.cloneDeep(origSettings)
    );

    const handleReset = useCallback(() => {
        setSettings(_.cloneDeep(origSettings));
    }, [origSettings]);

    useEffect(() => {
        handleReset();
        console.log('Loaded in');
    }, [handleReset, origSettings]);

    const handleSave = useCallback(() => {
        setSettings(settings);
    }, [settings]);

    const handleClose = () => {
        setOpen(false);
    };

    return (
        <Dialog open={open} fullWidth>
            <DialogTitle>Settings Editor</DialogTitle>
            <DialogContent dividers={true}>
                <DialogContentText paddingBottom={2}>General</DialogContentText>
                <Grid2 container spacing={1}>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Chat Warnings'}
                            tooltip={
                                'Enable in-game chat warnings to be broadcast to the active game'
                            }
                            enabled={settings.chat_warnings_enabled}
                            setEnabled={(chat_warnings_enabled) => {
                                setSettings({
                                    ...settings,
                                    chat_warnings_enabled
                                });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Kicker Enabled'}
                            tooltip={
                                'Enable the bot auto kick functionality when a match is found'
                            }
                            enabled={settings.kicker_enabled}
                            setEnabled={(kicker_enabled) => {
                                setSettings({ ...settings, kicker_enabled });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={12}>
                        <SettingsMultiSelect
                            label={'Kickable Tag Matches'}
                            tooltip={
                                'Only matches which also match these tags will trigger a kick or notification.'
                            }
                            values={settings.kick_tags}
                            setValues={(kick_tags) => {
                                setSettings({ ...settings, kick_tags });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Party Warnings Enabled'}
                            tooltip={
                                'Enable log messages to be broadcast to the lobby chat window'
                            }
                            enabled={settings.party_warnings_enabled}
                            setEnabled={(party_warnings_enabled) => {
                                setSettings({
                                    ...settings,
                                    party_warnings_enabled
                                });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Discord Presence Enabled'}
                            tooltip={
                                'Enable game status presence updates to your local discord client.'
                            }
                            enabled={settings.discord_presence_enabled}
                            setEnabled={(discord_presence_enabled) => {
                                setSettings({
                                    ...settings,
                                    discord_presence_enabled
                                });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Auto Launch Game On Start Up'}
                            tooltip={
                                'When enabled, upon launching bd, TF2 will also be launched at the same time'
                            }
                            enabled={settings.auto_launch_game}
                            setEnabled={(auto_launch_game) => {
                                setSettings({ ...settings, auto_launch_game });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Auto Close On Game Exit'}
                            tooltip={
                                'When enabled, upon the game existing, also shutdown bd.'
                            }
                            enabled={settings.auto_close_on_game_exit}
                            setEnabled={(auto_close_on_game_exit) => {
                                setSettings({
                                    ...settings,
                                    auto_close_on_game_exit
                                });
                            }}
                        />
                    </Grid2>

                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Enabled Debug Log'}
                            tooltip={
                                'When enabled, logs are written to bd.log in the application config root'
                            }
                            enabled={settings.debug_log_enabled}
                            setEnabled={(debug_log_enabled) => {
                                setSettings({ ...settings, debug_log_enabled });
                            }}
                        />
                    </Grid2>
                </Grid2>
                <DialogContentText paddingBottom={2}>
                    HTTP Service
                </DialogContentText>
                <Grid2 container>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Enable The HTTP Service*'}
                            tooltip={
                                'WARN: The HTTP service enabled the browser widget (this page) to function. You can only re-enable this ' +
                                'service by editing the config file manually'
                            }
                            enabled={settings.auto_close_on_game_exit}
                            setEnabled={(http_enabled) => {
                                setSettings({
                                    ...settings,
                                    http_enabled
                                });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsTextBox
                            label={'Listen Address (host:port)'}
                            tooltip={
                                'What address the http service will listen on. (The URL you are connected to right now). You should use localhost' +
                                'unless you know what you are doing as there is no authentication system.'
                            }
                            value={settings.http_listen_addr}
                            setValue={(http_listen_addr) => {
                                setSettings({ ...settings, http_listen_addr });
                            }}
                            validator={validatorAddress}
                        />
                    </Grid2>
                </Grid2>
                <DialogContentText paddingBottom={2}>
                    Steam Config
                </DialogContentText>
                <Grid2 container>
                    <Grid2 xs={6}>
                        <SettingsTextBox
                            label={'Steam ID'}
                            value={settings.steam_id}
                            setValue={(steam_id) => {
                                setSettings({ ...settings, steam_id });
                            }}
                            validator={validatorSteamID}
                            tooltip={
                                'You can choose one of the following formats: steam,steam3,steam64'
                            }
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsTextBox
                            label={'Steam API Key'}
                            value={settings.api_key}
                            secrets
                            validator={makeValidatorLength(32)}
                            setValue={(api_key) => {
                                setSettings({ ...settings, api_key });
                            }}
                            tooltip={'Your personal steam web api key'}
                        />
                    </Grid2>
                    <Grid2 xs={12}>
                        <SettingsTextBox
                            label={'Steam Root Directory'}
                            value={settings.steam_dir}
                            setValue={(steam_dir) => {
                                setSettings({ ...settings, steam_dir });
                            }}
                            tooltip={
                                'Location of your steam installation directory containing your userdata folder'
                            }
                        />
                    </Grid2>
                </Grid2>
                <DialogContentText paddingBottom={2}>
                    TF2 Config
                </DialogContentText>
                <Grid2 container>
                    <Grid2 xs={12}>
                        <SettingsTextBox
                            label={'TF2 Root Directory'}
                            value={settings.tf2_dir}
                            setValue={(tf2_dir) => {
                                setSettings({ ...settings, tf2_dir });
                            }}
                            tooltip={
                                'Path to your steamapps/common/Team Fortress 2/tf` Folder'
                            }
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'RCON Static Mode'}
                            tooltip={
                                'When enabled, rcon will always use the static port and password of 21212 / pazer_sux_lol. Otherwise these are generated randomly on game launch'
                            }
                            enabled={settings.chat_warnings_enabled}
                            setEnabled={(rcon_static) => {
                                setSettings({ ...settings, rcon_static });
                            }}
                        />
                    </Grid2>
                    <Grid2 xs={6}>
                        <SettingsCheckBox
                            label={'Generate Voice Bans'}
                            tooltip={
                                'WARN: This will overwrite your current ban list. Mutes the 200 most recent marked entries.'
                            }
                            enabled={settings.chat_warnings_enabled}
                            setEnabled={(voice_bans_enabled) => {
                                setSettings({
                                    ...settings,
                                    voice_bans_enabled
                                });
                            }}
                        />
                    </Grid2>
                </Grid2>
            </DialogContent>
            <DialogActions>
                <Button
                    onClick={handleClose}
                    startIcon={<CloseIcon />}
                    color={'error'}
                    // variant={'contained'}
                >
                    Cancel
                </Button>
                <Button
                    onClick={handleSave}
                    startIcon={<CheckIcon />}
                    color={'success'}
                    variant={'contained'}
                >
                    Save
                </Button>
            </DialogActions>
        </Dialog>
    );
};