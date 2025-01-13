import type { MDXComponents } from 'mdx/types'
import { ComponentPropsWithoutRef } from 'react';

import localFont from "next/font/local";
import { FaListOl, FaListUl, FaQuoteLeft } from 'react-icons/fa6';

const headingFont = localFont({
    src: "../public/fonts/Alliance2.otf",
    variable: "--font-heading",
});

const genericFont = localFont({
    src: "../public/fonts/Alliance1.otf",
    variable: "--font-generic",
});


type HeadingProps = ComponentPropsWithoutRef<'h1'>;

export function useMDXComponents(components: MDXComponents): MDXComponents {
  return {
    h1: (props: HeadingProps) => {
        return <h1 className={`${headingFont.className} font-bold text-4xl lg:text-5xl pt-12 mb-0`} {...props} />;
    },
    h2: (props: HeadingProps) => {
        return <h2 className={`${headingFont.className} font-bold text-3xl lg:text-4xl mt-8 mb-3`} {...props} />;
    },
    h3 : (props: HeadingProps) => {
        return <h3 className={`${headingFont.className} font-bold text-2xl lg:text-3xl mt-8 mb-3`} {...props} />;
    },
    h4 : (props: HeadingProps) => {
        return <h4 className={`${headingFont.className} font-bold text-xl lg:text-2xl mt-8 mb-3`} {...props} />;
    },
    h5 : (props: HeadingProps) => {
        return <h5 className={`${headingFont.className} font-bold text-lg lg:text-xl mt-8 mb-3`} {...props} />;
    },
    h6 : (props: HeadingProps) => {
        return <h6 className={`${headingFont.className} font-bold text-base lg:text-lg mt-8 mb-3`} {...props} />;
    },
    p : (props) => {
        return <p className={`${genericFont.className} text-lg lg:text-xl leading-snug`} {...props} />;
    },
    ul : ulComponent,
    ol : olComponent,
    li: (props) => {
        return <li className={`${genericFont.className} text-md lg:text-lg`} {...props} />;
    },
    blockquote : blockquoteComponent,
    code : (props) => {
        return <code className="dark:bg-black bg-white bg-opacity-20 text-brand p-1 rounded-md" {...props} />;
    },
    strong : (props) => {
        return <strong className="font-bold text-foreground" {...props} />;
    },
    ...components,
  }
}

const ulComponent = (props: {children: React.ReactNode}) => {    
    return (
        <div className="relative my-10 group-[.parent-list]:mt-0 group-[.parent-list]:mb-0">
            <div className={"absolute -left-2 -top-5 bg-brand p-2 rounded-xl group-[.parent-list]:hidden"}>
                <FaListUl className='text-background' size={'1.25rem'} />
            </div>
            <ul className={"list-disc dark:text-neutral-400 text-neutral-600 marker:text-brand marker:font-bold dark:bg-black bg-white dark:bg-opacity-20 pl-8 py-2 pt-6 pr-4 rounded-md space-y-1 border-brand border-l-4 group parent-list group-[.parent-list]:pt-2 group-[.parent-list]:pr-0 group-[.parent-list]:pl-4 group-[.parent-list]:border-l-0 group-[.parent-list]:bg-transparent"}>
                {props.children}
            </ul>
      </div>
    )
}

const olComponent = (props: {children: React.ReactNode}) => {
    return (
        <div className="relative my-10 group-[.parent-list]:mt-0 group-[.parent-list]:mb-0">
            <div className={"absolute -left-2 -top-5 bg-brand p-2 rounded-xl group-[.parent-list]:hidden"}>
                <FaListOl className='text-background' size={'1.25rem'} />
            </div>
            <ul className={"list-decimal dark:text-neutral-400 text-neutral-600 marker:text-brand marker:font-bold dark:bg-black bg-white dark:bg-opacity-20 pl-8 py-2 pt-6 pr-4 rounded-md space-y-1 border-brand border-l-4 group parent-list group-[.parent-list]:pt-2 group-[.parent-list]:pr-0 group-[.parent-list]:pl-4 group-[.parent-list]:border-l-0 group-[.parent-list]:bg-transparent"}>
                {props.children}
            </ul>
      </div>
    )
}

const blockquoteComponent = (props: {children: React.ReactNode}) => {
    return (
        <blockquote className="relative my-10 group-[.parent-list]:mt-0 group-[.parent-list]:mb-0">
            <div className={"absolute -left-2 -top-5 bg-brand p-2 rounded-xl group-[.parent-list]:hidden"}>
                <FaQuoteLeft className='text-background' size={'1.25rem'} />
            </div>
            <ol className={"dark:bg-black bg-white dark:bg-opacity-20 pl-2 py-2 pt-6 rounded-md space-y-1 border-brand border-l-4 group parent-list group-[.parent-list]:pt-2 group-[.parent-list]:ml-2 group-[.parent-list]:my-4 group-[.parent-list]:rounded-none group-[.parent-list]:border-l-2 group-[.parent-list]:border-opacity-60 group-[.parent-list]:bg-transparent"}>
                {props.children}
            </ol>
        </blockquote>
    )
};