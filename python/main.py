import matplotlib.pyplot as plt
import matplotlib
import numpy as np
import pandas as pd
from matplotlib.ticker import MultipleLocator

def plot_benchs(files,titles):

    fig, ax = plt.subplots(1,len(files),sharex=True, sharey=False)
    i=0
    for file in files:

        df = pd.read_csv(file, header=None)
        df.columns=["protocol","messages","size","time"]
        
        df["time"]=df["time"]/1000
        colors = {'quic':'red', 'tcp':'blue'}
    # colors=["red","blue"]
        #remove rows with 0 size and 0 time
        df=df[df["size"] != 0]
        df=df[df["size"] != 0]

        #plot the data x axis is the size of the message and y axis is the time, one line for
        #each protocol in diferent colors
        #interpolate dataframe dat

        
        protocol_groups=df.groupby(['protocol'])
        labels=[]
        positions=[]
        offset=-10
        bps=[]
        for key, grp in protocol_groups:
            labels.append(key)
            grup=grp.groupby(['size'])
            val=np.array(list(grup.groups.keys()))
            
            positions.append(val)
            
            
            data= [np.array(grup.get_group(x)["time"]) for x in grup.groups]
            color=colors[key]
            bp=ax[i].boxplot(data,showfliers=False,positions=val+offset,widths=10,patch_artist=True)
            for element in ['boxes', 'whiskers', 'fliers', 'medians', 'caps']:
                plt.setp(bp[element], color=color)
            for patch in bp['boxes']:
                patch.set(facecolor="white")
            bps.append(bp)
            offset+=20
            
        #get longer list of positions
    
        #set legend for all bps
       
        list_len = [len(i) for i in positions]
        positions= (positions[np.argmax(np.array(list_len))])
        ticks=np.arange(positions[0],positions[-1]+1,positions[1]-positions[0])
        
        ax[i].set_xticks(ticks)
        ax[i].set_xticklabels(ticks)
        #add legend and labels to each protocol

        #make y ticks every 50 ms
        start, end = ax[i].get_ylim()
        
        #yticks are from star to the next multiple of 5 of end
        end=end+10-end%10
        print(end)
        ticks = np.arange(0, end+1, end/10)
        


        ax[i].set_yticks(ticks)

        #each axes has a different y scale

        ax[i].set_yscale

        print(start,end)
       # ax[i].yaxis.set_ticks(np.arange(0, end, 50))
        ax[i].set_title(titles[i])
        
  
        


        i=i+1
    fig.supxlabel("Size (bytes)")
    fig.supylabel("Time (ms)")
    fig.suptitle("QUIC and TCP performance differences")
    fig.legend([bp["boxes"][0] for bp in bps],labels)
    fig.tight_layout()
    plt.show()
    plt.savefig(fname="bench_100.pdf",format="pdf")


def main():
  
    plot_benchs(["bench100_0drop.csv","bench_100_0drop_input.csv"],titles=["Test with 0% drop chance","Test with 0.5% drop chance"])




if __name__ == "__main__":
    main()
