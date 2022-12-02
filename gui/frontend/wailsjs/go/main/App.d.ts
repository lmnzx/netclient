// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {main} from '../models';
import {config} from '../models';
import {models} from '../models';

export function GoConnectToNetwork(arg1:string):Promise<any>;

export function GoDisconnectFromNetwork(arg1:string):Promise<any>;

export function GoGetKnownNetworks():Promise<Array<main.Network>>;

export function GoGetNetclientConfig():Promise<config.Config>;

export function GoGetNetwork(arg1:string):Promise<main.Network>;

export function GoJoinNetworkByToken(arg1:string):Promise<any>;

export function GoLeaveNetwork(arg1:string):Promise<any>;

export function GoParseAccessToken(arg1:string):Promise<models.AccessToken>;
