/* Minimal SVG icon set matching ux-ui/src/icons.jsx */
import { type SVGProps } from 'react';

type IconProps = SVGProps<SVGSVGElement> & { size?: number };

const icon = (path: string) =>
  function Icon({ size = 16, ...p }: IconProps) {
    return (
      <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" {...p}>
        <path d={path} />
      </svg>
    );
  };

const icon2 = (paths: string[]) =>
  function Icon({ size = 16, ...p }: IconProps) {
    return (
      <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" {...p}>
        {paths.map((d, i) => <path key={i} d={d} />)}
      </svg>
    );
  };

export const IGrid      = icon2(['M2 2h4v4H2z', 'M10 2h4v4h-4z', 'M2 10h4v4H2z', 'M10 10h4v4h-4z']);
export const IFolder    = icon2(['M1 3.5A1.5 1.5 0 012.5 2H6l2 2h5.5A1.5 1.5 0 0115 5.5v7a1.5 1.5 0 01-1.5 1.5h-11A1.5 1.5 0 011 12.5z']);
export const IUsers     = icon2(['M11 13v-1a3 3 0 00-3-3H5a3 3 0 00-3 3v1', 'M6.5 9a2.5 2.5 0 100-5 2.5 2.5 0 000 5z', 'M13 13v-1a3 3 0 00-2-2.83', 'M10.5 4a2.5 2.5 0 011.5 4.5']);
export const IKey       = icon2(['M10 2a4 4 0 100 8 4 4 0 000-8z', 'M6.2 9.8L2 14', 'M4 12l2 2', 'M2 14l2 2']);
export const IPlug      = icon2(['M9.5 2v3', 'M6.5 2v3', 'M4 5h8a2 2 0 010 4H4a2 2 0 010-4z', 'M8 9v5', 'M6 14h4']);
export const IShield    = icon('M8 1L2 4v4c0 3.3 2.5 6.4 6 7 3.5-.6 6-3.7 6-7V4L8 1z');
export const ISettings  = icon2(['M8 10a2 2 0 100-4 2 2 0 000 4z', 'M13.2 10a5.6 5.6 0 00.1-1 5.6 5.6 0 00-.1-1l1.4-1.1a.3.3 0 000-.4l-1.3-2.3a.3.3 0 00-.4-.1l-1.7.7a5.4 5.4 0 00-1.7-1L9.2 2.2A.3.3 0 009 2H6.6a.3.3 0 00-.3.2L6 3.8a5.4 5.4 0 00-1.7 1L2.6 4a.3.3 0 00-.4.1L.9 6.4a.3.3 0 000 .4L2.3 8a5.6 5.6 0 000 2L.9 11.1a.3.3 0 000 .4l1.3 2.3a.3.3 0 00.4.1l1.7-.7a5.4 5.4 0 001.7 1l.3 1.6a.3.3 0 00.3.2H9a.3.3 0 00.3-.2l.3-1.6a5.4 5.4 0 001.7-1l1.7.7a.3.3 0 00.4-.1l1.3-2.3a.3.3 0 000-.4L13.2 10z']);
export const IZap       = icon('M9 1L2 9h6l-1 6 7-8H8L9 1z');
export const ILayers    = icon2(['M8 1L1 5l7 4 7-4-7-4z', 'M1 9l7 4 7-4', 'M1 12l7 4 7-4']);
export const IRocket    = icon2(['M10.5 1.5A6 6 0 0114.5 5.5c0 4-4.5 8-6.5 9.5C6 13.5 1.5 9.5 1.5 5.5A6 6 0 015.5 1.5c1 2 2.5 3 2.5 3s1.5-1 2.5-3z', 'M8 7.5a1.5 1.5 0 100-3 1.5 1.5 0 000 3z', 'M3 11l-2 2', 'M13 11l2 2']);
export const IServer    = icon2(['M2 3.5A1.5 1.5 0 013.5 2h9A1.5 1.5 0 0114 3.5v3A1.5 1.5 0 0112.5 8h-9A1.5 1.5 0 012 6.5z', 'M2 9.5A1.5 1.5 0 013.5 8h9A1.5 1.5 0 0114 9.5v3a1.5 1.5 0 01-1.5 1.5h-9A1.5 1.5 0 012 12.5z', 'M5 5h.01', 'M5 11h.01']);
export const IBell      = icon2(['M8 1a5 5 0 015 5v3l1 2H2l1-2V6a5 5 0 015-5z', 'M6.5 14a1.5 1.5 0 003 0']);
export const ISearch    = icon2(['M7 12A5 5 0 107 2a5 5 0 000 10z', 'M11 11l3 3']);
export const IUser      = icon2(['M8 8a3 3 0 100-6 3 3 0 000 6z', 'M2 14a6 6 0 0112 0']);
export const IEdit      = icon2(['M11.5 2.5a2.12 2.12 0 013 3L5 15H2v-3L11.5 2.5z']);
export const ITrash     = icon2(['M2 4h12', 'M5 4V2h6v2', 'M3 4l1 10a1 1 0 001 1h6a1 1 0 001-1l1-10']);
export const IPlus      = icon('M8 2v12M2 8h12');
export const IMinus     = icon('M2 8h12');
export const ICheck     = icon('M2 8l4 4 8-8');
export const IX         = icon('M2 2l12 12M14 2L2 14');
export const IChevR     = icon('M6 3l5 5-5 5');
export const IChevL     = icon('M10 3L5 8l5 5');
export const IChevD     = icon('M3 6l5 5 5-5');
export const IArrL      = icon2(['M10 8H2', 'M5 5L2 8l3 3']);
export const IArrR      = icon2(['M6 8h8', 'M11 5l3 3-3 3']);
export const IHome      = icon2(['M1 7L8 1l7 6', 'M3 6v8h3.5v-4h3v4H13V6']);
export const IRefresh   = icon2(['M2 8A6 6 0 1114 8', 'M2 4v4h4']);
export const IFilter    = icon('M1 3h14M4 8h8M6.5 13h3');
export const IMore      = icon2(['M8 4.5a.5.5 0 110-1 .5.5 0 010 1z', 'M8 8.5a.5.5 0 110-1 .5.5 0 010 1z', 'M8 12.5a.5.5 0 110-1 .5.5 0 010 1z']);
export const IPlay      = icon('M4 2l10 6-10 6V2z');
export const IPause     = icon2(['M5 2v12', 'M11 2v12']);
export const IClock     = icon2(['M8 1a7 7 0 100 14A7 7 0 008 1z', 'M8 5v3.5L11 10']);
export const ICalendar  = icon2(['M2 5.5A1.5 1.5 0 013.5 4h9A1.5 1.5 0 0114 5.5v8a1.5 1.5 0 01-1.5 1.5h-9A1.5 1.5 0 012 13.5z', 'M5 2v4', 'M11 2v4', 'M2 8h12']);
export const ICopy      = icon2(['M11 1H3a1 1 0 00-1 1v10', 'M5 4h8a1 1 0 011 1v9a1 1 0 01-1 1H5a1 1 0 01-1-1V5a1 1 0 011-1z']);
export const ILink      = icon2(['M9 6.5a2.5 2.5 0 010 3.5L7.5 11.5a2.5 2.5 0 01-3.5-3.5L5 7', 'M7 9.5a2.5 2.5 0 010-3.5L8.5 4.5a2.5 2.5 0 013.5 3.5L11 9']);
export const ISync      = icon2(['M2 8A6 6 0 1114 8', 'M14 4v4h-4']);
export const IAlert     = icon2(['M8 1L15 14H1L8 1z', 'M8 6v4', 'M8 11.5v.5']);
export const IDrag      = icon2(['M5 4h.01', 'M5 8h.01', 'M5 12h.01', 'M11 4h.01', 'M11 8h.01', 'M11 12h.01']);
export const ITarget    = icon2(['M8 1a7 7 0 100 14A7 7 0 008 1z', 'M8 5a3 3 0 100 6 3 3 0 000-6z', 'M8 7.5v.5', 'M8 8h.5']);
export const IBook      = icon2(['M4 2h9a1 1 0 011 1v10a1 1 0 01-1 1H4a2 2 0 010-4h9', 'M4 14a2 2 0 010-4']);
export const IHeart     = icon('M8 13S2 9.5 2 5.5A3.5 3.5 0 018 3.5 3.5 3.5 0 0114 5.5C14 9.5 8 13 8 13z');
export const ICode      = icon2(['M5 4L1 8l4 4', 'M11 4l4 4-4 4', 'M9 2L7 14']);
export const IBranch    = icon2(['M4 2v10', 'M12 6a2 2 0 100 4 2 2 0 000-4z', 'M4 14a2 2 0 100-4 2 2 0 000 4z', 'M4 10a6 6 0 006-4']);
export const ITerminal  = icon2(['M3 11l3-3-3-3', 'M8 13h5']);
export const ICube      = icon2(['M14 11V5a1.4 1.4 0 00-.7-1.2L8.7 1.2a1.4 1.4 0 00-1.4 0L2.7 3.8A1.4 1.4 0 002 5v6a1.4 1.4 0 00.7 1.2l4.6 2.6a1.4 1.4 0 001.4 0l4.6-2.6A1.4 1.4 0 0014 11z', 'M2.2 4.6L8 8l5.8-3.4', 'M8 14.5V8']);
export const IGitCommit = icon2(['M8 11a3 3 0 100-6 3 3 0 000 6z', 'M1 8h4', 'M11 8h4']);
export const IUndo      = icon2(['M3 6a4 4 0 017-2l3 3', 'M13 7l-2-3h3']);
export const IRedo      = icon2(['M13 6a4 4 0 00-7-2L3 7', 'M3 7l2-3H2']);
export const IUp        = icon('M3 10l5-5 5 5');
export const IDown      = icon('M3 6l5 5 5-5');
