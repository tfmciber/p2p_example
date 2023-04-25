// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {peer} from '../models';
import {main} from '../models';

export function Add(arg1:string,arg2:peer.ID):Promise<void>;

export function AddDm(arg1:peer.ID):Promise<void>;

export function AddRendezvous(arg1:string):Promise<void>;

export function CancelRendezvous():Promise<void>;

export function Clear():Promise<void>;

export function DataChanged():Promise<void>;

export function DeleteChat(arg1:string):Promise<void>;

export function DeriveKey(arg1:Array<number>,arg2:Array<number>):Promise<Array<number>|Array<number>>;

export function DhtRoutine(arg1:boolean):Promise<void>;

export function EmitEvent(arg1:string,arg2:Array<any>):Promise<void>;

export function FakeUsers():Promise<Array<main.Users>>;

export function Get(arg1:string):Promise<Array<peer.ID>|boolean>;

export function GetData():Promise<{[key: string]: any}>;

export function GetKeys():Promise<Array<string>>;

export function GetPeerIDfromstring(arg1:string):Promise<peer.ID>;

export function GetRend():Promise<Array<string>>;

export function GetTimer(arg1:string):Promise<number>;

export function InitDHT():Promise<void>;

export function LeaveChat(arg1:string):Promise<void>;

export function ListChats():Promise<Array<string>>;

export function ListUsers():Promise<Array<main.Users>>;

export function MoveQueue():Promise<void>;

export function NewHost():Promise<string>;

export function NewID(arg1:string,arg2:string):Promise<void>;

export function OpenFileExplorer(arg1:string):Promise<void>;

export function OpenID(arg1:Array<number>,arg2:string):Promise<string>;

export function QueueFile(arg1:string,arg2:string):Promise<void>;

export function ReadKeys(arg1:string):Promise<Array<number>>;

export function Reconnect(arg1:string):Promise<void>;

export function SelectFiles():Promise<Array<main.PathFilename>>;

export function SendFile(arg1:string,arg2:string):Promise<void>;

export function SendTextHandler(arg1:string,arg2:string):Promise<boolean>;

export function SetPeers(arg1:string,arg2:Array<peer.ID>):Promise<void>;

export function SetTimer(arg1:string,arg2:number):Promise<void>;
